// Package proxy implements the reverse proxy Sharely uses to put an
// already-running local dev server (Vite, webpack-dev-server, etc.) onto
// the LAN, including WebSocket upgrades for HMR.
package proxy

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// New builds a reverse proxy to upstream (e.g. "http://127.0.0.1:5173").
// Plain HTTP requests are forwarded with httputil.ReverseProxy; requests
// carrying a WebSocket Upgrade header are bridged manually so HMR/live
// reload keeps working across the LAN.
func New(upstream string, quiet bool) (http.Handler, error) {
	target, err := url.Parse(upstream)
	if err != nil {
		return nil, err
	}

	rp := httputil.NewSingleHostReverseProxy(target)
	origDirector := rp.Director
	rp.Director = func(r *http.Request) {
		origDirector(r)
		r.Host = target.Host
	}
	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		if !quiet {
			log.Printf("proxy error: %v", err)
		}
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("The local dev server isn't responding."))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isWebsocketUpgrade(r) {
			proxyWebSocket(w, r, target)
			return
		}
		rp.ServeHTTP(w, r)
	}), nil
}

func isWebsocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

// proxyWebSocket hijacks the client connection and bridges raw bytes to a
// new TCP connection to the upstream, after replaying the original
// handshake request so the upstream's own WebSocket library completes the
// upgrade.
func proxyWebSocket(w http.ResponseWriter, r *http.Request, target *url.URL) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "WebSocket proxying isn't supported", http.StatusInternalServerError)
		return
	}

	upstreamConn, err := net.DialTimeout("tcp", target.Host, 5*time.Second)
	if err != nil {
		http.Error(w, "The local dev server isn't responding.", http.StatusBadGateway)
		return
	}

	if err := r.Write(upstreamConn); err != nil {
		upstreamConn.Close()
		http.Error(w, "Couldn't reach the local dev server.", http.StatusBadGateway)
		return
	}

	clientConn, _, err := hj.Hijack()
	if err != nil {
		upstreamConn.Close()
		return
	}

	done := make(chan struct{}, 2)
	go pipe(clientConn, upstreamConn, done)
	go pipe(upstreamConn, clientConn, done)
	<-done
}

func pipe(dst io.Writer, src io.Reader, done chan struct{}) {
	io.Copy(dst, src)
	if closer, ok := dst.(interface{ CloseWrite() error }); ok {
		closer.CloseWrite()
	}
	done <- struct{}{}
}
