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

	"sharely/internal/sharing"
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

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/"+s.ID+"/", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body, _ := io.ReadAll(rr.Result().Body)
	if string(body) != "<h1>hi</h1>" {
		t.Fatalf("unexpected body: %s", body)
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
	req := httptest.NewRequest(http.MethodGet, "/"+s.ID+"/", nil)
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
	req := httptest.NewRequest(http.MethodGet, "/"+s.ID+"/", nil)
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
	req := httptest.NewRequest(http.MethodGet, "/"+s.ID+"/", nil)
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
	req4 := httptest.NewRequest(http.MethodGet, "/"+s.ID+"/", nil)
	req4.AddCookie(cookies[0])
	h.ServeHTTP(rr4, req4)
	if rr4.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid session cookie, got %d", rr4.Code)
	}
}
