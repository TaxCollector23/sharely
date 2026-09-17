// Command sharely shares files, folders, and local sites over your network.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
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
	"golang.org/x/term"
)

const version = "0.2.2"

const daemonSubcommand = "__daemon"

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		runShare(nil)
		return
	}

	switch args[0] {
	case mdnsSubcommand:
		// Internal-only: see mdns_child.go for why this runs as its own
		// process instead of inside the daemon.
		runMDNSChild(args[1:])
	case daemonSubcommand:
		f, err := parseShareFlags(args[1:])
		if err != nil {
			fail(err.Error())
		}
		runDaemon(f)
	case "help", "-h", "--help":
		printHelp()
	case "version", "-v", "--version":
		fmt.Println("sharely " + version)
	case "list":
		runList()
	case "status":
		runList()
	case "dashboard":
		runDashboard()
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
	case "qr":
		if len(args) < 2 {
			fail("Usage: sharely qr <id>")
		}
		runQR(args[1])
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

func parseShareFlags(args []string) (shareFlags, error) {
	f := shareFlags{expires: "1h"}
	var positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--port":
			if i+1 >= len(args) {
				return f, fmt.Errorf("--port needs a number")
			}
			i++
			p, err := strconv.Atoi(args[i])
			if err != nil || p < 1 || p > 65535 {
				return f, fmt.Errorf("invalid port %q (use 1-65535)", args[i])
			}
			f.port = p
		case a == "--host":
			if i+1 >= len(args) {
				return f, fmt.Errorf("--host needs an IP address")
			}
			i++
			if ip := net.ParseIP(args[i]); ip == nil || ip.To4() == nil {
				return f, fmt.Errorf("invalid host %q (use an IPv4 address)", args[i])
			}
			f.host = args[i]
		case a == "--password":
			f.password = true
		case a == "--expires":
			if i+1 >= len(args) {
				return f, fmt.Errorf("--expires needs 15m, 1h, 4h, or until-stopped")
			}
			i++
			f.expires = normalizeExpires(args[i])
			if f.expires != "15m" && f.expires != "1h" && f.expires != "4h" && f.expires != "forever" {
				return f, fmt.Errorf("invalid expiration %q (use 15m, 1h, 4h, or until-stopped)", args[i])
			}
		case a == "--open":
			f.open = true
		case a == "--quiet":
			f.quiet = true
		case a == "--verbose":
			f.verbose = true
		case a == "--help" || a == "-h":
			printHelp()
			os.Exit(0)
		case strings.HasPrefix(a, "-"):
			return f, fmt.Errorf("unknown option %q\nRun `sharely help` to see available options.", a)
		default:
			positional = append(positional, a)
		}
	}
	if len(positional) > 1 {
		return f, fmt.Errorf("sharely accepts one file or folder at a time")
	}
	if len(positional) > 0 {
		f.target = positional[0]
	}
	return f, nil
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
	client := localHTTPClient(400 * time.Millisecond)
	resp, err := client.Get(controlBaseURL() + "/api/status")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// ---- sharely <target> ---------------------------------------------------

func runShare(args []string) {
	f, err := parseShareFlags(args)
	if err != nil {
		fail(err.Error())
	}
	// Resolve in the invoking process so bad paths fail immediately and
	// relative paths are never interpreted against the daemon's directory.
	f.target = targetOrCwd(f.target)
	if _, err := sharing.Resolve(f.target); err != nil {
		fail(err.Error())
	}

	if !daemonReachable() {
		if err := startDaemon(f); err != nil {
			fail(err.Error())
		}
	}
	runAsClient(f)
}

type shareResult struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Target        string  `json:"target"`
	Type          string  `json:"type"`
	URL           string  `json:"url"`
	PrimaryURL    string  `json:"primaryUrl"`
	Remaining     string  `json:"remaining"`
	HasPassword   bool    `json:"hasPassword"`
	Duration      string  `json:"duration"`
	NetworkAddr   string  `json:"networkAddress"`
	LocalHostname string  `json:"localHostname"`
	Password      *string `json:"-"`
}

func runAsClient(f shareFlags) {
	body, _ := json.Marshal(map[string]string{
		"target":   targetOrCwd(f.target),
		"duration": f.expires,
		"password": clientPassword(f),
	})
	client := localHTTPClient(3 * time.Second)
	resp, err := client.Post(controlBaseURL()+"/api/shares", "application/json", bytes.NewReader(body))
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
			candidate := "http://127.0.0.1:" + strconv.Itoa(port) + "/?share=" + url.QueryEscape(res.ID)
			if httpReachable(candidate) {
				fallbackURL = candidate
			}
		}
	}

	printReady(f, res, reachable, fallbackURL)
	if f.verbose {
		passwordLabel := "disabled"
		if res.HasPassword {
			passwordLabel = "enabled"
		}
		printVerboseDetails([][2]string{
			{"Share ID", res.ID},
			{"Type", res.Type},
			{"Target", res.Target},
			{"Content address", res.NetworkAddr},
			{"Local hostname", nonEmpty(res.LocalHostname, "(unavailable, using LAN address)")},
			{"Password", passwordLabel},
			{"Duration", res.Duration},
		})
	}
	if f.open {
		openBrowser(res.PrimaryURL)
	}
}

// httpReachable does a real GET (not just a TCP dial) so a link is only
// ever called "ready" once something actually answers on it — matching
// Sharely's rule that a pretty-but-broken URL is worse than an honest
// warning.
func httpReachable(url string) bool {
	client := localHTTPClient(900 * time.Millisecond)
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode < 500
}

// localHTTPClient deliberately ignores HTTP_PROXY. Sharely only probes local
// addresses; routing a private LAN URL through a corporate proxy can produce a
// false "not responding" warning even while the share works perfectly.
func localHTTPClient(timeout time.Duration) http.Client {
	return http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: nil,
		},
	}
}

func clientPassword(f shareFlags) string {
	if !f.password {
		return ""
	}
	return promptPassword()
}

func promptPassword() string {
	// Always stderr, never stdout — stdout is reserved for the share link
	// itself (especially in --quiet mode, where scripts pipe it straight
	// into another command and can't tolerate a stray prompt line).
	fmt.Fprint(os.Stderr, "Set a password for this share: ")
	if term.IsTerminal(int(os.Stdin.Fd())) {
		pw, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			fail("Sharely couldn't read the password.")
		}
		if len(pw) == 0 {
			fail("Password cannot be empty.")
		}
		return string(pw)
	}
	var pw string
	if _, err := fmt.Scanln(&pw); err != nil || pw == "" {
		fail("Password cannot be empty.")
	}
	return pw
}

// targetOrCwd resolves the user's target argument to an absolute path in
// THIS process, before it ever crosses a process boundary. This matters
// specifically for client mode (runAsClient): the daemon that ultimately
// creates the share may be running with a completely different working
// directory (it could have been started from another terminal, sharing a
// different folder, minutes or hours earlier), so a relative path like
// "report.pdf" must never be sent over the wire as-is — the daemon would
// resolve it against ITS OWN cwd and fail with a confusing "does not
// exist" error even though the file is sitting right in front of the user.
func targetOrCwd(t string) string {
	if t == "" {
		if wd, err := os.Getwd(); err == nil {
			return wd
		}
		return "."
	}
	if abs, err := filepath.Abs(t); err == nil {
		return abs
	}
	return t
}

// startDaemon launches the long-running server independently so `sharely
// start` returns the prompt immediately and the share survives terminal exit.
func startDaemon(f shareFlags) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("Sharely couldn't locate its executable: %w", err)
	}
	args := []string{daemonSubcommand}
	if f.host != "" {
		args = append(args, "--host", f.host)
	}
	if f.port != 0 {
		args = append(args, "--port", strconv.Itoa(f.port))
	}

	dir, err := config.Dir()
	if err != nil {
		return fmt.Errorf("Sharely couldn't create its local state directory: %w", err)
	}
	logPath := filepath.Join(dir, "daemon.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("Sharely couldn't open its daemon log: %w", err)
	}
	defer logFile.Close()

	cmd := exec.Command(exe, args...)
	cmd.Stdin = nil
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = daemonProcessAttributes()
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("Sharely couldn't start in the background: %w", err)
	}
	pid := strconv.Itoa(cmd.Process.Pid)
	_ = os.WriteFile(filepath.Join(dir, "daemon.pid"), []byte(pid+"\n"), 0o600)
	_ = cmd.Process.Release()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if daemonReachable() {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("Sharely didn't finish starting. See %s", logPath)
}

func runDaemon(f shareFlags) {
	if dir, err := config.Dir(); err == nil {
		defer os.Remove(filepath.Join(dir, "daemon.pid"))
	}
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
	contentAddr := net.JoinHostPort(lanIP, strconv.Itoa(contentPort))
	listenHost := lanIP
	if f.host == "" && lanIP != "127.0.0.1" {
		// Binding wildcard avoids interface-specific loopback quirks on macOS
		// and still advertises the concrete LAN address to other devices.
		listenHost = "0.0.0.0"
	}
	listenAddr := net.JoinHostPort(listenHost, strconv.Itoa(contentPort))

	// Clients use this stable loopback endpoint. Do not silently move the
	// daemon to another port: that would leave a healthy but undiscoverable
	// background process behind when the preferred port is occupied.
	controlAddr := net.JoinHostPort(config.ControlHost, strconv.Itoa(config.ControlPortDefault))

	// Advertise "sharely.local" over mDNS, but only ever hand it to the user
	// as the link's hostname once we've confirmed THIS machine can actually
	// resolve it. A responder can bind successfully while the OS's own
	// resolver still can't find ".local" names (no nss-mdns/Avahi on plain
	// Linux, multicast blocked by a firewall, a restrictive network
	// profile, etc.) — printing an unverified hostname is exactly the
	// "pretty but broken link" failure Sharely must never produce.
	//
	// The responder itself runs as a separate OS process (startMDNSAdvertiser,
	// see mdns_child.go): it parses arbitrary real mDNS traffic from every
	// device on the LAN, and isolating it means nothing it does — including
	// crashing outright — can ever take the actual file-sharing server down
	// with it.
	localName := ""
	var mdnsProc *mdnsProcess
	if ip := parseIP(lanIP); ip != nil {
		if proc, err := startMDNSAdvertiser(lanIP, contentPort); err == nil {
			mdnsProc = proc
			// Give the child a brief moment to bind before checking whether
			// this machine can resolve it — startup is near-instant in
			// practice, and Verify()'s own timeout absorbs the rest.
			time.Sleep(700 * time.Millisecond)
			if runtime.GOOS == "darwin" || discovery.Verify() {
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
	if listenHost != "0.0.0.0" && lanIP != "127.0.0.1" && lanIP != "localhost" {
		loopbackFallback = "127.0.0.1:" + strconv.Itoa(contentPort)
	}

	srv := &server.Server{
		Manager:              mgr,
		ContentAddr:          listenAddr,
		ContentHost:          contentAddr,
		ControlAddr:          controlAddr,
		LocalName:            localName,
		IfaceLabel:           ifaceLabel,
		Version:              version,
		Quiet:                f.quiet,
		DashboardDir:         dashboardDir,
		LoopbackFallbackAddr: loopbackFallback,
	}
	shutdownRequested := make(chan struct{}, 1)
	srv.OnShutdown = func() {
		select {
		case shutdownRequested <- struct{}{}:
		default:
		}
	}
	if err := srv.Start(); err != nil {
		fail("Sharely couldn't start: " + err.Error())
	}

	// Block until interrupted, then clean up every share and both listeners.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sig:
	case <-shutdownRequested:
	}
	mdnsProc.Stop()
	mgr.StopAll()
	srv.Shutdown()
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
	fmt.Printf("  Dashboard  http://%s\n", net.JoinHostPort(config.ControlHost, strconv.Itoa(config.ControlPortDefault)))
	fmt.Println("  Open, copy, or scan your share from the dashboard.")
	fmt.Printf("\n  Share link  %s\n\n", res.PrimaryURL)
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
	if host, _, ok := splitHostPortInt(res.NetworkAddr); ok && host == "127.0.0.1" {
		fmt.Println()
		fmt.Println("⚠ No local network was found. This share is only available on this computer.")
		fmt.Println("  Connect to Wi-Fi or Ethernet, then run `sharely doctor`.")
		return
	}
	if !reachable {
		fmt.Println()
		if fallbackURL != "" {
			fmt.Println("⚠ Sharely couldn't verify the network link from this computer.")
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

// printVerboseDetails prints the technical details --verbose promises,
// which until now the flag silently ignored — it parsed but did nothing.
func printVerboseDetails(rows [][2]string) {
	fmt.Println()
	fmt.Println("  Details")
	width := 0
	for _, r := range rows {
		if len(r[0]) > width {
			width = len(r[0])
		}
	}
	for _, r := range rows {
		fmt.Printf("    %-*s  %s\n", width, r[0], r[1])
	}
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
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "..", "web", "dashboard", "dist"))
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
