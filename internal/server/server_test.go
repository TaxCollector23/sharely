package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/TaxCollector23/sharely/internal/sharing"
)

func newTestHandler(t *testing.T) (*ContentHandler, *sharing.Manager, string) {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(dir, "index.html"), []byte("<h1>hi</h1>"), 0o644)
	os.WriteFile(filepath.Join(dir, ".env"), []byte("TOKEN=x"), 0o644)
	os.MkdirAll(filepath.Join(dir, "sub"), 0o755)
	os.WriteFile(filepath.Join(dir, "sub", "a.txt"), []byte("a"), 0o644)

	m := sharing.NewManager()
	h := NewContentHandler(m, true)
	return h, m, dir
}

func TestServeDirectoryAndFile(t *testing.T) {
	h, m, dir := newTestHandler(t)
	s, err := m.Create(sharing.CreateOptions{TargetArg: dir, RootDir: dir, EntryPoint: "index.html", Type: sharing.TypeWebsite, Duration: sharing.Duration1Hour})
	if err != nil {
		t.Fatal(err)
	}
	pretty := httptest.NewRecorder()
	h.ServeHTTP(pretty, httptest.NewRequest(http.MethodGet, "/"+s.ID+"/", nil))
	if pretty.Code != http.StatusSeeOther || pretty.Header().Get("Location") != "/?share="+s.ID {
		t.Fatalf("expected friendly path to enter SPA root, got %d %q", pretty.Code, pretty.Header().Get("Location"))
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?share="+s.ID, nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body, _ := io.ReadAll(rr.Result().Body)
	if !strings.Contains(string(body), "<h1>hi</h1>") {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestWebsiteHTMLKeepsAssetsInsideSharePrefix(t *testing.T) {
	h, m, dir := newTestHandler(t)
	markup := `<!doctype html><html><head><link rel="stylesheet" href="/assets/app.css"><script src="/assets/app.js"></script><link href="//cdn.example.com/x.css"></head><body></body></html>`
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(markup), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := m.Create(sharing.CreateOptions{TargetArg: dir, RootDir: dir, EntryPoint: "index.html", Type: sharing.TypeWebsite, Duration: sharing.Duration1Hour})
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/?share="+s.ID, nil))
	body := rr.Body.String()
	if !strings.Contains(body, `href="/`+s.ID+`/assets/app.css"`) || !strings.Contains(body, `src="/`+s.ID+`/assets/app.js"`) {
		t.Fatalf("root-relative assets were not scoped to the share: %s", body)
	}
	if !strings.Contains(body, `href="//cdn.example.com/x.css"`) {
		t.Fatalf("protocol-relative URL was unexpectedly rewritten: %s", body)
	}
}

func TestRootShareSelectionSupportsSPAAssets(t *testing.T) {
	h, m, dir := newTestHandler(t)
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "app.js"), []byte("document.body.dataset.ready='yes'"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := m.Create(sharing.CreateOptions{TargetArg: dir, RootDir: dir, EntryPoint: "index.html", Type: sharing.TypeWebsite, Duration: sharing.Duration1Hour})
	if err != nil {
		t.Fatal(err)
	}

	page := httptest.NewRecorder()
	h.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/?share="+s.ID, nil))
	if page.Code != http.StatusOK || len(page.Result().Cookies()) == 0 {
		t.Fatalf("expected selected share and cookie, got %d", page.Code)
	}

	assetReq := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	assetReq.AddCookie(page.Result().Cookies()[0])
	asset := httptest.NewRecorder()
	h.ServeHTTP(asset, assetReq)
	if asset.Code != http.StatusOK {
		t.Fatalf("expected SPA asset through selected share, got %d", asset.Code)
	}
	if got := asset.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/javascript") {
		t.Fatalf("unexpected JavaScript MIME type: %q", got)
	}

	routeReq := httptest.NewRequest(http.MethodGet, "/innovation", nil)
	routeReq.AddCookie(page.Result().Cookies()[0])
	route := httptest.NewRecorder()
	h.ServeHTTP(route, routeReq)
	if route.Code != http.StatusOK || !strings.Contains(route.Body.String(), "<h1>hi</h1>") {
		t.Fatalf("expected SPA route to fall back to index, got %d: %s", route.Code, route.Body.String())
	}
}

func TestServeBlocksTraversalAndSensitiveFiles(t *testing.T) {
	h, m, dir := newTestHandler(t)
	s, _ := m.Create(sharing.CreateOptions{TargetArg: dir, RootDir: dir, Type: sharing.TypeDirectory, Duration: sharing.Duration1Hour})

	cases := []string{
		"/" + s.ID + "/../../../etc/passwd",
		"/" + s.ID + "/.env",
	}
	for _, path := range cases {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		h.ServeHTTP(rr, req)
		if rr.Code == http.StatusOK {
			t.Errorf("expected %s to be blocked, got 200", path)
		}
	}
}

func TestExpiredShareServesGone(t *testing.T) {
	h, m, dir := newTestHandler(t)
	s, _ := m.Create(sharing.CreateOptions{TargetArg: dir, RootDir: dir, Type: sharing.TypeDirectory, Duration: sharing.Duration15Min})
	past := time.Now().Add(-time.Second)
	s.ExpiresAt = &past

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?share="+s.ID, nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusGone {
		t.Fatalf("expected 410 Gone for expired share, got %d", rr.Code)
	}
}

func TestStoppedShareIsUnreachable(t *testing.T) {
	h, m, dir := newTestHandler(t)
	s, _ := m.Create(sharing.CreateOptions{TargetArg: dir, RootDir: dir, Type: sharing.TypeDirectory, Duration: sharing.DurationForever})
	m.Stop(s.ID)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?share="+s.ID, nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusGone {
		t.Fatalf("expected 410 Gone for stopped share, got %d", rr.Code)
	}
}

func TestUnknownShareIs404(t *testing.T) {
	h, _, _ := newTestHandler(t)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/nonexistent/", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestPasswordProtectedShareRequiresAuth(t *testing.T) {
	h, m, dir := newTestHandler(t)
	s, err := m.Create(sharing.CreateOptions{
		TargetArg: dir, RootDir: dir, EntryPoint: "index.html", Type: sharing.TypeWebsite,
		Duration: sharing.Duration1Hour, Password: "hunter2",
	})
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?share="+s.ID, nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without auth, got %d", rr.Code)
	}

	// Wrong password.
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/"+s.ID+"/_auth", strings.NewReader("password=wrong"))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong password, got %d", rr2.Code)
	}

	// Correct password yields a session cookie that then grants access.
	rr3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPost, "/"+s.ID+"/_auth", strings.NewReader("password=hunter2"))
	req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect after correct password, got %d", rr3.Code)
	}
	cookies := rr3.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected a session cookie to be set")
	}

	rr4 := httptest.NewRecorder()
	req4 := httptest.NewRequest(http.MethodGet, "/?share="+s.ID, nil)
	req4.AddCookie(cookies[0])
	h.ServeHTTP(rr4, req4)
	if rr4.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid session cookie, got %d", rr4.Code)
	}
}

func TestShutdownRequiresJSONAndSignals(t *testing.T) {
	called := make(chan struct{}, 1)
	h := &APIHandler{Manager: sharing.NewManager(), Shutdown: func() { called <- struct{}{} }}

	bad := httptest.NewRecorder()
	h.Routes().ServeHTTP(bad, httptest.NewRequest(http.MethodPost, "/api/shutdown", nil))
	if bad.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415 for a form-compatible request, got %d", bad.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/shutdown", nil)
	req.Header.Set("Content-Type", "application/json")
	good := httptest.NewRecorder()
	h.Routes().ServeHTTP(good, req)
	if good.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", good.Code)
	}
	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("shutdown callback was not called")
	}
}

func TestFriendlyHostnameKeepsContentPort(t *testing.T) {
	m := sharing.NewManager()
	dir := t.TempDir()
	s, err := m.Create(sharing.CreateOptions{TargetArg: dir, RootDir: dir, Type: sharing.TypeDirectory, Duration: sharing.Duration1Hour})
	if err != nil {
		t.Fatal(err)
	}
	h := &APIHandler{Manager: m, ContentHost: "10.0.0.8:4821", LocalName: "sharely.local"}
	got := h.primaryURL(s)
	want := "http://sharely.local:4821/" + s.ID + "/"
	if got != want {
		t.Fatalf("friendly URL = %q, want %q", got, want)
	}
	if dtoURL := h.toDTO(s).PrimaryURL; dtoURL != got {
		t.Fatalf("API URL = %q, QR target = %q", dtoURL, got)
	}
}
