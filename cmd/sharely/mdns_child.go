package main

import (
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/TaxCollector23/sharely/internal/discovery"
)

// mdnsSubcommand is an unadvertised, internal-only CLI entry point. It
// exists so the mDNS responder — which parses arbitrary multicast UDP
// packets from every Bonjour/mDNS device on the real LAN (printers,
// Chromecasts, HomeKit accessories, other laptops, ...), none of which
// Sharely controls or can fully trust — runs in its own OS process rather
// than inside the daemon that also serves the actual shared content.
//
// Go's net/http already recovers a panic inside any single request
// handler without taking down the process, but a panic inside a raw
// background goroutine owned by a third-party library (as the mDNS
// responder's packet-handling loop is) is NOT recoverable from outside
// that goroutine, and would otherwise crash the entire sharely daemon —
// silently taking the whole share down along with the merely-cosmetic
// "sharely.local" hostname feature. Isolating it here means the worst
// mDNS can do is stop resolving the friendly hostname; the actual file
// server keeps running on its LAN IP no matter what happens to it.
const mdnsSubcommand = "__mdns_advertise"

func runMDNSChild(args []string) {
	if len(args) < 1 {
		os.Exit(1)
	}
	ip := net.ParseIP(args[0])
	if ip == nil {
		os.Exit(1)
	}

	stop, err := discovery.Advertise(ip)
	if err != nil {
		os.Exit(1)
	}
	defer stop()

	// Block until either signaled to stop, or our stdin pipe closes —
	// which happens automatically at the OS level the moment the parent
	// daemon exits for ANY reason, including a crash or SIGKILL. This is
	// what keeps this helper from ever being left running as an orphan.
	stdinClosed := make(chan struct{})
	go func() {
		io.Copy(io.Discard, os.Stdin)
		close(stdinClosed)
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	select {
	case <-stdinClosed:
	case <-sig:
	}
}

// mdnsProcess supervises the child process started by startMDNSAdvertiser.
type mdnsProcess struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser
}

func (m *mdnsProcess) Stop() {
	if m == nil {
		return
	}
	m.stdin.Close()
	done := make(chan struct{})
	go func() {
		m.cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		m.cmd.Process.Kill()
	}
}

// startMDNSAdvertiser launches the isolated mDNS responder for ip and
// returns a handle to stop it. Any failure to launch is non-fatal to the
// caller — mDNS is always optional, LAN-IP sharing must keep working
// regardless.
func startMDNSAdvertiser(ip string) (*mdnsProcess, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(exe, mdnsSubcommand, ip)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &mdnsProcess{cmd: cmd, stdin: stdin}, nil
}
