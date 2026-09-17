// Package discovery advertises a friendly "sharely.local" mDNS hostname
// when the local network supports it. It is best-effort only: Sharely must
// keep working perfectly with a plain LAN IP if mDNS is unavailable.
package discovery

import (
	"context"
	"io"
	"log"
	"net"
	"time"

	"github.com/hashicorp/mdns"
	"github.com/miekg/dns"
)

// hostZone answers mDNS A-record queries for sharely.local with ip.
type hostZone struct {
	ip net.IP
}

func (z *hostZone) Records(q dns.Question) []dns.RR {
	if q.Qtype != dns.TypeA && q.Qtype != dns.TypeANY {
		return nil
	}
	rr := &dns.A{
		Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
		A:   z.ip,
	}
	return []dns.RR{rr}
}

// Advertise starts an mDNS responder for "sharely.local" -> ip and returns a
// shutdown func. It returns an error if the responder could not bind
// (e.g. multicast blocked); callers should fall back to the LAN IP.
func Advertise(ip net.IP) (stop func(), err error) {
	// Other Bonjour devices often send DNSSEC/NSEC records that older mDNS
	// parsers cannot decode. The responder should remain best-effort and
	// never flood Sharely's terminal with parser noise for packets it can
	// safely ignore.
	server, err := mdns.NewServer(&mdns.Config{
		Zone:   &hostZone{ip: ip},
		Logger: log.New(io.Discard, "", 0),
	})
	if err != nil {
		return nil, err
	}
	return func() { server.Shutdown() }, nil
}

// Verify does a best-effort check that sharely.local actually resolves,
// used by `sharely doctor`. It never blocks for long.
func Verify() bool {
	resolver := &net.Resolver{}
	done := make(chan bool, 1)
	go func() {
		_, err := resolver.LookupHost(context.Background(), "sharely.local")
		done <- err == nil
	}()
	select {
	case ok := <-done:
		return ok
	case <-time.After(800 * time.Millisecond):
		return false
	}
}
