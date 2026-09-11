// Package network finds the active LAN interface and a free port to bind,
// so the rest of Sharely never has to think about interfaces or sockets.
package network

import (
	"fmt"
	"net"
	"strings"
)

// Interface describes a candidate network interface.
type Interface struct {
	Name string
	IP   string
}

var ignoredPrefixes = []string{"lo", "docker", "veth", "br-", "utun", "tun", "tap", "bridge", "awdl", "llw", "anpi", "gif", "stf"}

// ActiveLAN returns the best-guess LAN interface Sharely should bind to:
// the first non-loopback, non-virtual interface carrying a private IPv4
// address and currently up.
func ActiveLAN() (*Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var candidates []Interface
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		lname := strings.ToLower(iface.Name)
		skip := false
		for _, p := range ignoredPrefixes {
			if strings.HasPrefix(lname, p) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipNet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipNet.IP.To4()
			if ip4 == nil || !ip4.IsPrivate() {
				continue
			}
			candidates = append(candidates, Interface{Name: iface.Name, IP: ip4.String()})
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no active LAN interface found")
	}
	// Prefer typical Wi-Fi/Ethernet naming (en0 on macOS, wlan0/eth0 on Linux).
	for _, c := range candidates {
		if c.Name == "en0" || c.Name == "eth0" || c.Name == "wlan0" {
			return &c, nil
		}
	}
	return &candidates[0], nil
}

// FreePort finds an available TCP port, starting at preferred and scanning
// forward if it's taken.
func FreePort(host string, preferred int) (int, error) {
	for p := preferred; p < preferred+200; p++ {
		l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, p))
		if err == nil {
			l.Close()
			return p, nil
		}
	}
	// Fall back to an OS-assigned port.
	l, err := net.Listen("tcp", host+":0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// Reachable reports whether host:port currently accepts connections.
func Reachable(host string, port int) bool {
	c, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 500_000_000)
	if err != nil {
		return false
	}
	c.Close()
	return true
}
