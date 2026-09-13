package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/TaxCollector23/sharely/internal/discovery"
	"github.com/TaxCollector23/sharely/internal/network"
	"github.com/TaxCollector23/sharely/internal/qr"
)

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

func execCommand(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Run()
}

func apiGet(path string, out any) error {
	client := localHTTPClient(2 * time.Second)
	resp, err := client.Get(controlBaseURL() + path)
	if err != nil {
		return fmt.Errorf("not running")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s", strings.TrimSpace(string(data)))
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("request failed (%d)", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func apiPost(path string, body any, out any) error {
	client := localHTTPClient(2 * time.Second)
	var reader io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reader = strings.NewReader(string(data))
	}
	req, _ := http.NewRequest(http.MethodPost, controlBaseURL()+path, reader)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("not running")
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s", strings.TrimSpace(string(data)))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func apiDelete(path string, out any) error {
	client := localHTTPClient(2 * time.Second)
	req, _ := http.NewRequest(http.MethodDelete, controlBaseURL()+path, nil)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("not running")
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s", strings.TrimSpace(string(data)))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func runList() {
	if !daemonReachable() {
		fmt.Println("Nothing is being shared.")
		fmt.Println()
		fmt.Println("Run `sharely start` from a folder to get started.")
		return
	}
	var shares []struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Status     string `json:"status"`
		Remaining  string `json:"remaining"`
		PrimaryURL string `json:"primaryUrl"`
	}
	if err := apiGet("/api/shares", &shares); err != nil {
		fail(err.Error())
	}
	active := shares[:0]
	for _, s := range shares {
		if s.Status == "running" {
			active = append(active, s)
		}
	}
	if len(active) == 0 {
		fmt.Println("Nothing is being shared.")
		return
	}
	fmt.Println("Active shares")
	fmt.Println()
	for _, s := range active {
		fmt.Printf("%-10s %-14s %-28s %s\n", s.ID, truncate(s.Name, 14), s.PrimaryURL, s.Remaining)
	}
}

func runDashboard() {
	if !daemonReachable() {
		fail("Sharely isn't running. Start a share first with `sharely start`.")
	}
	url := controlBaseURL()
	openBrowser(url)
	fmt.Println("Opened " + url)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func runStop(id string) {
	if !daemonReachable() {
		fail("Sharely isn't running.")
	}
	if err := apiPost("/api/shares/"+id+"/stop", nil, nil); err != nil {
		fail("No active share named \"" + id + "\".")
	}
	fmt.Println("Sharing stopped.")
}

func runStopAll() {
	if !daemonReachable() {
		fmt.Println("Nothing is being shared.")
		return
	}
	var shares []struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	apiGet("/api/shares", &shares)
	n := 0
	for _, s := range shares {
		if s.Status == "running" {
			apiPost("/api/shares/"+s.ID+"/stop", nil, nil)
			n++
		}
	}
	if n == 0 {
		apiPost("/api/shutdown", nil, nil)
		fmt.Println("Nothing is being shared.")
		return
	}
	apiPost("/api/shutdown", nil, nil)
	fmt.Println("Sharing stopped.")
}

func runLogs(id string) {
	if !daemonReachable() {
		fail("Sharely isn't running.")
	}
	var logs []struct {
		Time   time.Time `json:"time"`
		Method string    `json:"method"`
		Path   string    `json:"path"`
		Status int       `json:"status"`
	}
	if err := apiGet("/api/shares/"+id+"/logs", &logs); err != nil {
		fail(err.Error())
	}
	fmt.Println(id)
	fmt.Println()
	if len(logs) == 0 {
		fmt.Println("No requests yet.")
		return
	}
	for _, l := range logs {
		fmt.Printf("%s  %-5s  %3d  %s\n", l.Time.Format("15:04:05"), l.Method, l.Status, l.Path)
	}
}

// runQR reprints the link and terminal QR code for an already-running
// share — handy after scrolling past it, after `--quiet`, or from a second
// terminal that didn't create the share itself.
func runQR(id string) {
	if !daemonReachable() {
		fail("Sharely isn't running.")
	}
	var share struct {
		PrimaryURL string `json:"primaryUrl"`
		Remaining  string `json:"remaining"`
	}
	if err := apiGet("/api/shares/"+id, &share); err != nil {
		fail("No active share named \"" + id + "\".")
	}
	fmt.Println()
	fmt.Printf("  %s\n\n", share.PrimaryURL)
	fmt.Printf("  %s\n\n", share.Remaining)
	art, err := qr.Terminal(share.PrimaryURL)
	if err == nil {
		fmt.Println(art)
	}
	fmt.Println("  Scan to open")
}

func runDoctor() {
	fmt.Println("Sharely diagnostics")
	fmt.Println()

	iface, err := network.ActiveLAN()
	if err != nil {
		fmt.Println("✗ No active LAN interface found")
		fmt.Println("  Your share will only be reachable from this machine.")
	} else {
		fmt.Printf("✓ Network interface detected (%s)\n", iface.Name)
		fmt.Printf("✓ LAN address available: %s\n", iface.IP)
	}

	port, err := network.FreePort("0.0.0.0", 4821)
	if err == nil {
		fmt.Printf("✓ Port available (%d)\n", port)
	} else {
		fmt.Println("✗ Could not find a free port")
	}

	if daemonReachable() {
		fmt.Println("✓ Sharely daemon reachable")
		var status struct {
			ActiveCount int    `json:"activeCount"`
			Network     string `json:"network"`
			LocalHost   string `json:"localHost"`
		}
		if apiGet("/api/status", &status) == nil {
			fmt.Printf("  %d active share(s) on %s\n", status.ActiveCount, status.Network)
		}
	} else {
		fmt.Println("- No Sharely daemon currently running")
	}

	if discovery.Verify() {
		fmt.Println("✓ Local hostname discovery available (sharely.local)")
	} else {
		fmt.Println("⚠ Local hostname discovery unavailable")
		fmt.Println("  Your share will use the LAN address instead.")
	}
}

func printHelp() {
	fmt.Println(`Share files, folders, and local sites over your network.

Usage:
  sharely start [path]     Share a folder (defaults to the current one)
  sharely [path]           Same thing — "start" is just the explicit form
  sharely list             See what's currently shared
  sharely status           Alias for list
  sharely dashboard        Open the local control center
  sharely qr <id>          Reprint the link and QR code for a share
  sharely stop <id>        Stop one share
  sharely stop-all         Stop everything
  sharely logs <id>        Recent requests for a share
  sharely doctor           Check your network setup
  sharely version

Examples:
  cd my-project && sharely start
  sharely index.html
  sharely ./project
  sharely report.pdf

Options:
  --port <n>       Use a specific port
  --host <ip>      Bind to a specific address
  --password       Require a password to view
  --expires <t>    15m, 1h, 4h, or until-stopped (default 1h)
  --open           Open the link in your browser
  --quiet          Print only the link
  --verbose        Show technical logs`)
}
