package server

import (
	"context"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TaxCollector23/sharely/internal/sharing"
	dashboard "github.com/TaxCollector23/sharely/web/dashboard"
)

// Server wires the two listeners Sharely runs: the LAN-facing content
// server (untrusted origin, serves shares) and the loopback-only control
// server (dashboard + JSON API). Keeping them on separate ports/origins is
// deliberate: it's what stops arbitrary JS in a shared page from ever being
// able to reach the control API. See content.go / api.go for detail.
type Server struct {
	Manager *sharing.Manager

	ContentAddr string // e.g. "192.168.1.42:4821"
	ContentHost string // advertised host:port used in links and API responses
	ControlAddr string // e.g. "127.0.0.1:47732"
	LocalName   string
	IfaceLabel  string
	Version     string
	Quiet       bool
	OnShutdown  func()

	DashboardDir string // path to built dashboard static assets, if any

	// LoopbackFallbackAddr, when set (e.g. "127.0.0.1:4821"), makes the
	// content server also listen here in addition to ContentAddr. Binding
	// the LAN interface can silently fail to be self-reachable on some
	// setups (restrictive firewalls, certain container/VPN networking) even
	// though it works fine for other devices — this loopback listener is
	// Sharely's guarantee that "does the link work on this machine" always
	// has a real answer, never left to chance. Left empty when ContentAddr
	// is already loopback.
	LoopbackFallbackAddr string

	contentSrv         *http.Server
	contentLoopbackSrv *http.Server
	controlSrv         *http.Server
}

func (s *Server) Start() error {
	contentHandler := NewContentHandler(s.Manager, s.Quiet)
	s.contentSrv = &http.Server{Addr: s.ContentAddr, Handler: contentHandler}

	if s.LoopbackFallbackAddr != "" {
		s.contentLoopbackSrv = &http.Server{Addr: s.LoopbackFallbackAddr, Handler: contentHandler}
	}

	api := &APIHandler{
		Manager:     s.Manager,
		ContentHost: nonEmptyHost(s.ContentHost, s.ContentAddr),
		LocalName:   s.LocalName,
		Interface:   s.IfaceLabel,
		StartedAt:   time.Now(),
		Version:     s.Version,
		Shutdown:    s.OnShutdown,
	}

	mux := http.NewServeMux()
	mux.Handle("/api/", api.Routes())
	mux.Handle("/", dashboardHandler(s.DashboardDir))
	s.controlSrv = &http.Server{Addr: s.ControlAddr, Handler: withNoCORS(mux)}

	errCh := make(chan error, 3)
	go func() {
		serveTCP4(s.contentSrv, s.ContentAddr, errCh)
	}()
	if s.contentLoopbackSrv != nil {
		go func() {
			// Best-effort: if this address is somehow already taken, the
			// LAN listener above is still the primary content server and
			// Sharely keeps working — this is purely a fallback.
			serveTCP4(s.contentLoopbackSrv, s.LoopbackFallbackAddr, errCh)
		}()
	}
	go func() {
		serveTCP4(s.controlSrv, s.ControlAddr, errCh)
	}()

	go s.sweep()

	select {
	case err := <-errCh:
		return err
	case <-time.After(150 * time.Millisecond):
		return nil
	}
}

// serveTCP4 makes the transport explicit. On macOS, net/http may resolve a
// wildcard TCP listener to IPv6 first; some local networks then accept an IPv4
// connection and immediately reset it. Sharely advertises IPv4 LAN links, so
// a TCP4 listener keeps the URL and the socket family aligned.
func serveTCP4(srv *http.Server, addr string, errCh chan<- error) {
	l, err := net.Listen("tcp4", addr)
	if err != nil {
		errCh <- err
		return
	}
	if err := srv.Serve(l); err != nil && err != http.ErrServerClosed {
		errCh <- err
	}
}

func nonEmptyHost(preferred, fallback string) string {
	if preferred != "" {
		return preferred
	}
	return fallback
}

func (s *Server) sweep() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.Manager.Sweep(24 * time.Hour)
	}
}

func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if s.contentSrv != nil {
		s.contentSrv.Shutdown(ctx)
	}
	if s.contentLoopbackSrv != nil {
		s.contentLoopbackSrv.Shutdown(ctx)
	}
	if s.controlSrv != nil {
		s.controlSrv.Shutdown(ctx)
	}
}

// withNoCORS makes explicit that the control API never opts into
// cross-origin access; browsers will refuse any request to it that
// originates from the (different-origin) content server without this.
func withNoCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

// dashboardHandler serves the built dashboard SPA, falling back to
// index.html for client-side routes. It prefers an on-disk directory when
// one is explicitly configured (handy for `npm run dev` iteration), and
// otherwise serves the copy embedded into the sharely binary at build time
// (see web/dashboard/embed.go) so a single compiled executable never
// depends on a separate asset directory being present at install time.
func dashboardHandler(dir string) http.Handler {
	embedded, embedErr := fs.Sub(dashboard.Dist, dashboard.DistSubdir)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if dir != "" {
			full := filepath.Join(dir, filepath.Clean("/"+r.URL.Path))
			if info, err := os.Stat(full); err == nil && !info.IsDir() {
				http.ServeFile(w, r, full)
				return
			}
			index := filepath.Join(dir, "index.html")
			if _, err := os.Stat(index); err == nil {
				http.ServeFile(w, r, index)
				return
			}
		}

		if embedErr == nil {
			reqPath := strings.TrimPrefix(r.URL.Path, "/")
			if reqPath == "" {
				reqPath = "index.html"
			}
			if f, err := embedded.Open(reqPath); err == nil {
				f.Close()
				http.ServeFileFS(w, r, embedded, reqPath)
				return
			}
			if f, err := embedded.Open("index.html"); err == nil {
				f.Close()
				http.ServeFileFS(w, r, embedded, "index.html")
				return
			}
		}

		writeHTML(w, http.StatusOK, dashboardMissingPage())
	})
}

func dashboardMissingPage() string {
	return `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>Sharely dashboard</title><style>` + pageCSS + `</style></head>
<body class="sharely-page"><main class="sharely-card">
<h1>Dashboard not built</h1>
<p class="muted">Run the dashboard's build step and point Sharely at the output directory.</p>
</main></body></html>`
}

// ResolveLoopback finds a free loopback port for the control server.
func ResolveLoopback(host string, preferred int) (string, error) {
	for p := preferred; p < preferred+50; p++ {
		l, err := net.Listen("tcp", net.JoinHostPort(host, itoaLocal(p)))
		if err == nil {
			l.Close()
			return net.JoinHostPort(host, itoaLocal(p)), nil
		}
	}
	return "", os.ErrExist
}

func itoaLocal(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
