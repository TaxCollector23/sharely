// Command sharely shares files, folders, and local sites over your network.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/TaxCollector23/sharely/internal/config"
	"github.com/TaxCollector23/sharely/internal/discovery"
	"github.com/TaxCollector23/sharely/internal/network"
	"github.com/TaxCollector23/sharely/internal/qr"
	"github.com/TaxCollector23/sharely/internal/server"
	"github.com/TaxCollector23/sharely/internal/sharing"
)

const version = "0.1.0"

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		runShare(nil)
		return
	}

	switch args[0] {
	case "help", "-h", "--help":
		printHelp()
	case "version", "-v", "--version":
		fmt.Println("sharely " + version)
	case "list":
		runList()
	case "stop":
		if len(args) < 2 {
			fail("Usage: sharely stop <id>")
		}
		runStop(args[1])
	case "stop-all":
		runStopAll()
	case "doctor":
		runDoctor()
	case "logs":
		if len(args) < 2 {
			fail("Usage: sharely logs <id>")
		}
		runLogs(args[1])
	case "start":
		// `sharely start` is the explicit, discoverable spelling of
		// `sharely [path]`: cd into a folder, run `sharely start`, get a
		// link. Both spellings do exactly the same thing.
		runShare(args[1:])
	default:
		runShare(args)
	}
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

// ---- flags -----------------------------------------------------------

type shareFlags struct {
	target   string
	port     int
	host     string
	password bool
	expires  string
	open     bool
	quiet    bool
	verbose  bool
}

func parseShareFlags(args []string) shareFlags {
	f := shareFlags{expires: "1h"}
	var positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--port" && i+1 < len(args):
			i++
			p, err := strconv.Atoi(args[i])
			if err == nil {
				f.port = p
			}
		case a == "--host" && i+1 < len(args):
			i++
			f.host = args[i]
		case a == "--password":
			f.password = true
		case a == "--expires" && i+1 < len(args):
			i++
			f.expires = normalizeExpires(args[i])
		case a == "--open":
			f.open = true
		case a == "--quiet":
			f.quiet = true
		case a == "--verbose":
			f.verbose = true
		default:
			positional = append(positional, a)
		}
	}
	if len(positional) > 0 {
		f.target = positional[0]
	}
	return f
}

func normalizeExpires(v string) string {
	switch strings.ToLower(v) {
	case "15m", "15min", "15minutes":
		return "15m"
	case "1h", "1hour", "hour":
		return "1h"
	case "4h", "4hours":
		return "4h"
	case "until-stopped", "forever", "never":
		return "forever"
	default:
		return v
	}
}

// ---- daemon discovery --------------------------------------------------

func controlBaseURL() string {
	return "http://" + config.ControlHost + ":" + strconv.Itoa(config.ControlPortDefault)
}

func daemonReachable() bool {
	client := http.Client{Timeout: 400 * time.Millisecond}
	resp, err := client.Get(controlBaseURL() + "/api/status")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// ---- sharely <target> ---------------------------------------------------

func runShare(args []string) {
	f := parseShareFlags(args)

	if daemonReachable() {
		runAsClient(f)
		return
	}
	runAsDaemon(f)
}

type shareResult struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	URL         string  `json:"url"`
	PrimaryURL  string  `json:"primaryUrl"`
	Remaining   string  `json:"remaining"`
	HasPassword bool    `json:"hasPassword"`
	Duration    string  `json:"duration"`
	NetworkAddr string  `json:"networkAddress"`
	Password    *string `json:"-"`
}

func runAsClient(f shareFlags) {
	body, _ := json.Marshal(map[string]string{
		"target":   targetOrCwd(f.target),
		"duration": f.expires,
		"password": clientPassword(f),
	})
	resp, err := http.Post(controlBaseURL()+"/api/shares", "application/json", bytes.NewReader(body))
	if err != nil {
		fail("Sharely is running, but couldn't be reached. Try `sharely doctor`.")
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		var e struct{ Error string }
		json.Unmarshal(data, &e)
		if e.Error != "" {
			fail(e.Error)
		}
		fail("Sharely couldn't create that share.")
	}
	var res shareResult
	json.Unmarshal(data, &res)

	reachable := httpReachable(res.URL)
	fallbackURL := ""
	if !reachable {
		if _, port, ok := splitHostPortInt(res.NetworkAddr); ok {
			candidate := "http://127.0.0.1:" + strconv.Itoa(port) + "/" + res.ID + "/"
			if httpReachable(candidate) {
				fallbackURL = candidate
			}
		}
	}

	printReady(f, res, reachable, fallbackURL)
	if f.open {
		openBrowser(res.PrimaryURL)
	}
}

// httpReachable does a real GET (not just a TCP dial) so a link is only
// ever called "ready" once something actually answers on it — matching
// Sharely's rule that a pretty-but-broken URL is worse than an honest
// warning.
func httpReachable(url string) bool {
	client := http.Client{Timeout: 700 * time.Millisecond}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode < 500
}

func clientPassword(f shareFlags) string {
	if !f.password {
		return ""
	}
	return promptPassword()
}

func promptPassword() string {
	fmt.Print("Set a password for this share: ")
	var pw string
	fmt.Scanln(&pw)
	return pw
}

func targetOrCwd(t string) string {
	if t == "" {
		if wd, err := os.Getwd(); err == nil {
			return wd
		}
		return "."
	}
	return t
}

func runAsDaemon(f shareFlags) {
	mgr := sharing.NewManager()

	iface, err := network.ActiveLAN()
	lanIP := "127.0.0.1"
	ifaceLabel := "unknown"
	if err == nil {
		lanIP = iface.IP
		ifaceLabel = iface.Name
	} else if !f.quiet {
		fmt.Println("Sharely couldn't find a usable network connection. Sharing will only work on this machine.")
	}
	if f.host != "" {
		lanIP = f.host
	}

	contentPort := config.ContentPortDefault
	if f.port != 0 {
		contentPort = f.port
	}
	contentPort, err = network.FreePort(lanIP, contentPort)
	if err != nil {
		fail("Sharely couldn't open a network port. Try again or run `sharely doctor`.")
	}
	contentAddr := lanIP + ":" + strconv.Itoa(contentPort)

	controlAddr, err := server.ResolveLoopback(config.ControlHost, config.ControlPortDefault)
	if err != nil {
		fail("Sharely couldn't start its control server.")
	}

	// Advertise "sharely.local" over mDNS, but only ever hand it to the user
	// as the link's hostname once we've confirmed THIS machine can actually
	// resolve it. A responder can bind successfully while the OS's own
	// resolver still can't find ".local" names (no nss-mdns/Avahi on plain
	// Linux, multicast blocked by a firewall, a restrictive network
	// profile, etc.) — printing an unverified hostname is exactly the
	// "pretty but broken link" failure Sharely must never produce.
	localName := ""
	var stopDiscovery func()
	if ip := parseIP(lanIP); ip != nil {
		if stop, err := discovery.Advertise(ip); err == nil {
			// Keep the responder running either way — a phone or another
			// computer on the LAN may resolve "sharely.local" fine even
			// when this machine's own resolver can't. We only gate
			// whether WE trust it enough to print as the primary link.
			stopDiscovery = stop
			if discovery.Verify() {
				localName = config.LocalHostname
			}
		}
	}

	dashboardDir := os.Getenv("SHARELY_DASHBOARD_DIR")
	if dashboardDir == "" {
		if p := findDashboardDir(); p != "" {
			dashboardDir = p
		}
	}

	loopbackFallback := ""
	if lanIP != "127.0.0.1" && lanIP != "localhost" {
		loopbackFallback = "127.0.0.1:" + strconv.Itoa(contentPort)
	}

	srv := &server.Server{
		Manager:              mgr,
		ContentAddr:          contentAddr,
		ControlAddr:          controlAddr,
		LocalName:            localName,
		IfaceLabel:           ifaceLabel,
		Version:              version,
		Quiet:                f.quiet,
		DashboardDir:         dashboardDir,
		LoopbackFallbackAddr: loopbackFallback,
	}
	if err := srv.Start(); err != nil {
		fail("Sharely couldn't start: " + err.Error())
	}

	resolved, rerr := sharing.Resolve(f.target)
	if rerr != nil {
		fail(rerr.Error())
	}
	password := clientPassword(f)
	s, err := mgr.Create(sharing.CreateOptions{
		TargetArg:     targetOrCwd(f.target),
		RootDir:       resolved.RootDir,
		EntryPoint:    resolved.EntryPoint,
		Type:          resolved.Type,
		ProxyUpstream: resolved.ProxyUpstream,
		Duration:      sharing.Duration(f.expires),
		Password:      password,
	})
	if err != nil {
		fail(err.Error())
	}

	res := shareResult{
		ID:          s.ID,
		Name:        baseName(s.TargetArg),
		Type:        string(s.Type),
		URL:         "http://" + contentAddr + "/" + s.ID + "/",
		PrimaryURL:  "http://" + contentAddr + "/" + s.ID + "/",
		Remaining:   s.RemainingLabel(),
		HasPassword: s.HasPassword,
		Duration:    string(s.Duration),
		NetworkAddr: contentAddr,
	}
	if localName != "" {
		res.PrimaryURL = "http://" + localName + "/" + s.ID + "/"
	}

	// Actually fetch the link we're about to print — not just probe the
	// socket — before calling it ready. If it doesn't come back cleanly,
	// check the loopback listener, which always works on this machine
	// regardless of LAN/firewall/container quirks, and offer it as a
	// guaranteed-working backup instead of leaving the user with a link
	// that looks fine but silently doesn't load.
	primaryReachable := httpReachable(res.URL)
	fallbackURL := ""
	if !primaryReachable && loopbackFallback != "" {
		candidate := "http://" + loopbackFallback + "/" + s.ID + "/"
		if httpReachable(candidate) {
			fallbackURL = candidate
		}
	}

	printReady(f, res, primaryReachable, fallbackURL)
	if !f.quiet {
		fmt.Printf("\n  Dashboard   http://%s\n", controlAddr)
	}
	if f.open {
		openBrowser(res.PrimaryURL)
	}

	// Block until interrupted, then clean up every share and both listeners.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	if stopDiscovery != nil {
		stopDiscovery()
	}
	mgr.StopAll()
	srv.Shutdown()
	if !f.quiet {
		fmt.Println("\nSharing stopped.")
	}
}

func baseName(p string) string {
	p = strings.TrimSuffix(p, "/")
	if p == "." || p == "" {
		wd, err := os.Getwd()
		if err == nil {
			parts := strings.Split(wd, "/")
			return parts[len(parts)-1]
		}
		return "current directory"
	}
	parts := strings.Split(p, "/")
	return parts[len(parts)-1]
}

func printReady(f shareFlags, res shareResult, reachable bool, fallbackURL string) {
	if f.quiet {
		if !reachable && fallbackURL != "" {
			fmt.Println(fallbackURL)
			return
		}
		fmt.Println(res.PrimaryURL)
		return
	}
	typeLabel := map[string]string{
		"file": "file", "directory": "folder", "website": "site", "proxy": "dev server",
	}[res.Type]

	fmt.Println()
	fmt.Println("Sharely")
	fmt.Println()
	fmt.Printf("Sharing  %s\n\n", res.Name)
	fmt.Printf("✓ Your %s is ready.\n\n", nonEmpty(typeLabel, "content"))
	fmt.Printf("  %s\n\n", res.PrimaryURL)
	fmt.Printf("  %s\n", res.Remaining)
	if res.HasPassword {
		fmt.Println("  Password protected")
	}
	fmt.Println()
	art, err := qr.Terminal(res.PrimaryURL)
	if err == nil {
		fmt.Println(art)
	}
	fmt.Println("  Scan to open")
	if !reachable {
		fmt.Println()
		if fallbackURL != "" {
			fmt.Println("⚠ That link isn't responding from other devices on this network.")
			fmt.Printf("  It works on this computer at:  %s\n", fallbackURL)
			fmt.Println("  Run `sharely doctor` to check why other devices can't reach it.")
		} else {
			fmt.Println("⚠ This link isn't responding. Run `sharely doctor` to check your network setup.")
		}
	}
}

func nonEmpty(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

// splitHostPortInt parses "host:port" for a reachability check. ok is false
// (not an error) when addr isn't in that shape, so callers can treat an
// unparseable address as "nothing to verify" rather than "unreachable".
func splitHostPortInt(addr string) (host string, port int, ok bool) {
	idx := strings.LastIndex(addr, ":")
	if idx == -1 {
		return "", 0, false
	}
	p, err := strconv.Atoi(addr[idx+1:])
	if err != nil {
		return "", 0, false
	}
	return addr[:idx], p, true
}

func parseIP(s string) []byte {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return nil
	}
	out := make([]byte, 4)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || n > 255 {
			return nil
		}
		out[i] = byte(n)
	}
	return out
}

func findDashboardDir() string {
	candidates := []string{
		"web/dashboard/dist",
		"./web/dashboard/dist",
	}
	exe, err := os.Executable()
	if err == nil {
		candidates = append(candidates, exe+"/../web/dashboard/dist")
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	return ""
}

func openBrowser(url string) {
	var cmd string
	var args []string
	switch {
	case fileExists("/usr/bin/open"):
		cmd, args = "open", []string{url}
	case fileExists("/usr/bin/xdg-open"):
		cmd, args = "xdg-open", []string{url}
	default:
		return
	}
	execCommand(cmd, args...)
}
