package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/TaxCollector23/sharely/internal/mimekind"
	"github.com/TaxCollector23/sharely/internal/proxy"
	"github.com/TaxCollector23/sharely/internal/security"
	"github.com/TaxCollector23/sharely/internal/sharing"
	"golang.org/x/net/html"
)

// ContentHandler serves the LAN-facing side of Sharely: /<shareID>/<path>.
// This is the origin that untrusted, user-shared content lives on, so it
// must never expose the control API (see internal/server/api.go, which is
// only ever bound to loopback on a different port).
type ContentHandler struct {
	Manager *sharing.Manager
	Quiet   bool

	proxiesMu sync.Mutex
	proxies   map[string]http.Handler
}

func NewContentHandler(m *sharing.Manager, quiet bool) *ContentHandler {
	return &ContentHandler{Manager: m, Quiet: quiet, proxies: make(map[string]http.Handler)}
}

func (h *ContentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, rest := splitShareID(r.URL.Path)
	if selected := r.URL.Query().Get("share"); r.URL.Path == "/" && selected != "" {
		id = selected
		rest = "/"
		http.SetCookie(w, &http.Cookie{
			Name:     "sharely_selected",
			Value:    selected,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	} else if _, exists := h.Manager.Get(id); !exists {
		if cookie, err := r.Cookie("sharely_selected"); err == nil {
			if _, selectedExists := h.Manager.Get(cookie.Value); selectedExists {
				id = cookie.Value
				rest = r.URL.Path
			}
		}
	}
	if id == "" {
		writeHTML(w, http.StatusOK, landingPage())
		return
	}

	s, ok := h.Manager.Get(id)
	if !ok {
		writeHTML(w, http.StatusNotFound, notFoundPage())
		return
	}
	if s.Status() != sharing.StatusRunning {
		writeHTML(w, http.StatusGone, expiredPage())
		return
	}

	if rest == "/_auth" && r.Method == http.MethodPost {
		h.handleAuth(w, r, s)
		return
	}
	if !isAuthenticated(r, s) {
		renderPasswordPage(w, id, false)
		return
	}

	rec := &statusRecorder{ResponseWriter: w, status: 200}
	h.serveShareContent(rec, r, s, rest)
	s.Touch(visitorKey(r), r.Method, rest, rec.status)
}

func (h *ContentHandler) handleAuth(w http.ResponseWriter, r *http.Request, s *sharing.Share) {
	r.ParseForm()
	password := r.Form.Get("password")
	if !checkPassword(s, password) {
		renderPasswordPage(w, s.ID, true)
		return
	}
	setAuthCookie(w, s)
	http.Redirect(w, r, "/?share="+url.QueryEscape(s.ID), http.StatusSeeOther)
}

func (h *ContentHandler) serveShareContent(w http.ResponseWriter, r *http.Request, s *sharing.Share, rest string) {
	if s.Type == sharing.TypeProxy {
		h.serveProxy(w, r, s)
		return
	}

	decodedRest, err := url.PathUnescape(rest)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// A single shared file/website ignores any sub-path beyond its own
	// entry point request for "/" so `sharely index.html` just works.
	requestRel := strings.TrimPrefix(decodedRest, "/")
	if requestRel == "" && s.EntryPoint != "" {
		requestRel = s.EntryPoint
	}

	if security.IsSensitive(requestRel) {
		writeHTML(w, http.StatusNotFound, notFoundPage())
		return
	}

	full, err := security.Resolve(s.RootDir, requestRel)
	if err != nil {
		writeHTML(w, http.StatusForbidden, notFoundPage())
		return
	}

	info, err := os.Stat(full)
	if err != nil {
		if s.Type == sharing.TypeWebsite && r.Method == http.MethodGet && filepath.Ext(requestRel) == "" {
			indexPath := filepath.Join(s.RootDir, nonEmptyEntryPoint(s.EntryPoint))
			if fileExists(indexPath) {
				h.serveWebsiteHTML(w, s, indexPath)
				return
			}
		}
		writeHTML(w, http.StatusNotFound, notFoundPage())
		return
	}

	if info.IsDir() {
		h.serveDirectory(w, r, s, full, requestRel)
		return
	}

	mimeType, _ := mimekind.Detect(full)
	if s.Type == sharing.TypeWebsite && (strings.EqualFold(filepath.Ext(full), ".html") || strings.EqualFold(filepath.Ext(full), ".htm")) {
		h.serveWebsiteHTML(w, s, full)
		return
	}
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, full)
}

func nonEmptyEntryPoint(entry string) string {
	if entry != "" {
		return entry
	}
	return "index.html"
}

func (h *ContentHandler) serveDirectory(w http.ResponseWriter, r *http.Request, s *sharing.Share, dirPath, requestRel string) {
	// Missing trailing slash would break relative links from the listing.
	if !strings.HasSuffix(r.URL.Path, "/") {
		http.Redirect(w, r, r.URL.Path+"/", http.StatusMovedPermanently)
		return
	}
	if s.EntryPoint != "" && requestRel == "" {
		// Directory shares with a detected index just serve it directly.
		indexPath := filepath.Join(dirPath, "index.html")
		if fileExists(indexPath) {
			h.serveWebsiteHTML(w, s, indexPath)
			return
		}
	}
	if s.Type == sharing.TypeDirectory || s.Type == sharing.TypeWebsite {
		indexPath := filepath.Join(dirPath, "index.html")
		if fileExists(indexPath) {
			h.serveWebsiteHTML(w, s, indexPath)
			return
		}
	}
	urlBase := "/" + s.ID + "/" + strings.TrimSuffix(requestRel, "/")
	urlBase = strings.TrimSuffix(urlBase, "/")
	html, err := renderBrowser(s.ID, filepath.Base(s.TargetArg), urlBase, dirPath)
	if err != nil {
		writeHTML(w, http.StatusInternalServerError, notFoundPage())
		return
	}
	writeHTML(w, http.StatusOK, html)
}

// serveWebsiteHTML rewrites root-relative asset links so a built site can
// live below Sharely's /<share-id>/ URL. Vite and similar tools commonly
// emit /assets/app.js, which would otherwise escape the share prefix and
// leave the page with only its HTML title and a blank body.
func (h *ContentHandler) serveWebsiteHTML(w http.ResponseWriter, s *sharing.Share, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		writeHTML(w, http.StatusNotFound, notFoundPage())
		return
	}
	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
		return
	}
	prefix := "/" + s.ID + "/"
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for i := range n.Attr {
				a := &n.Attr[i]
				if a.Key != "src" && a.Key != "href" && a.Key != "action" && a.Key != "poster" {
					continue
				}
				if strings.HasPrefix(a.Val, "/") && !strings.HasPrefix(a.Val, "//") {
					a.Val = prefix + strings.TrimPrefix(a.Val, "/")
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := html.Render(w, doc); err != nil {
		writeHTML(w, http.StatusInternalServerError, notFoundPage())
	}
}

func (h *ContentHandler) serveProxy(w http.ResponseWriter, r *http.Request, s *sharing.Share) {
	h.proxiesMu.Lock()
	p, ok := h.proxies[s.ID]
	if !ok {
		var err error
		p, err = proxy.New(s.ProxyUpstream, h.Quiet)
		if err != nil {
			h.proxiesMu.Unlock()
			writeHTML(w, http.StatusBadGateway, notFoundPage())
			return
		}
		h.proxies[s.ID] = p
	}
	h.proxiesMu.Unlock()
	// Strip the /<id> prefix so the upstream sees the path it expects.
	r2 := r.Clone(r.Context())
	r2.URL.Path = "/" + strings.TrimPrefix(strings.TrimPrefix(r.URL.Path, "/"+s.ID), "/")
	if r2.URL.Path == "" {
		r2.URL.Path = "/"
	}
	p.ServeHTTP(w, r2)
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

// splitShareID pulls the share ID off the front of a request path and
// returns the remainder (including its leading slash, possibly empty).
func splitShareID(p string) (id, rest string) {
	trimmed := strings.TrimPrefix(p, "/")
	if trimmed == "" {
		return "", ""
	}
	idx := strings.Index(trimmed, "/")
	if idx == -1 {
		return trimmed, ""
	}
	return trimmed[:idx], trimmed[idx:]
}

func visitorKey(r *http.Request) string {
	ip := r.RemoteAddr
	if h, _, err := splitHostPort(ip); err == nil {
		ip = h
	}
	sum := sha256.Sum256([]byte(ip + "|" + r.UserAgent()))
	return hex.EncodeToString(sum[:8])
}

func splitHostPort(addr string) (string, string, error) {
	idx := strings.LastIndex(addr, ":")
	if idx == -1 {
		return addr, "", nil
	}
	return addr[:idx], addr[idx+1:], nil
}

func landingPage() string {
	return `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>Sharely</title><style>` + pageCSS + `</style></head>
<body class="sharely-page"><main class="sharely-card">
<h1>Sharely</h1>
<p class="muted">This machine is running Sharely. Ask for a share link to see what's being shared.</p>
</main></body></html>`
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
