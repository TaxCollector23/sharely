package sharing

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManagerCreateAssignsUniqueNames(t *testing.T) {
	m := NewManager()
	seen := map[string]bool{}
	for i := 0; i < 20; i++ {
		s, err := m.Create(CreateOptions{TargetArg: ".", Type: TypeDirectory, RootDir: t.TempDir(), Duration: Duration1Hour})
		if err != nil {
			t.Fatal(err)
		}
		if seen[s.ID] {
			t.Fatalf("duplicate share ID generated: %s", s.ID)
		}
		seen[s.ID] = true
	}
}

func TestExpirationDoesNotChangeID(t *testing.T) {
	m := NewManager()
	s, err := m.Create(CreateOptions{TargetArg: ".", Type: TypeDirectory, RootDir: t.TempDir(), Duration: Duration15Min})
	if err != nil {
		t.Fatal(err)
	}
	id := s.ID
	s.SetExpiration(Duration4Hour)
	if s.ID != id {
		t.Fatalf("changing expiration must not change the share ID")
	}
	if s.Duration != Duration4Hour {
		t.Fatalf("expected duration to update")
	}
}

func TestForeverHasNoExpiry(t *testing.T) {
	m := NewManager()
	s, _ := m.Create(CreateOptions{TargetArg: ".", Type: TypeDirectory, RootDir: t.TempDir(), Duration: DurationForever})
	if s.ExpiresAt != nil {
		t.Fatalf("expected nil ExpiresAt for 'until stopped' shares")
	}
	if s.Status() != StatusRunning {
		t.Fatalf("expected running status")
	}
}

func TestExpiredShareReportsExpired(t *testing.T) {
	m := NewManager()
	s, _ := m.Create(CreateOptions{TargetArg: ".", Type: TypeDirectory, RootDir: t.TempDir(), Duration: Duration15Min})
	past := time.Now().Add(-time.Minute)
	s.ExpiresAt = &past
	if s.Status() != StatusExpired {
		t.Fatalf("expected expired status, got %s", s.Status())
	}
}

func TestStopIsTerminal(t *testing.T) {
	m := NewManager()
	s, _ := m.Create(CreateOptions{TargetArg: ".", Type: TypeDirectory, RootDir: t.TempDir(), Duration: DurationForever})
	m.Stop(s.ID)
	if s.Status() != StatusStopped {
		t.Fatalf("expected stopped status")
	}
}

func TestStopAllOnlyCountsRunning(t *testing.T) {
	m := NewManager()
	a, _ := m.Create(CreateOptions{TargetArg: ".", Type: TypeDirectory, RootDir: t.TempDir(), Duration: DurationForever})
	m.Create(CreateOptions{TargetArg: ".", Type: TypeDirectory, RootDir: t.TempDir(), Duration: DurationForever})
	m.Stop(a.ID)
	if n := m.StopAll(); n != 1 {
		t.Fatalf("expected StopAll to report 1 newly-stopped share, got %d", n)
	}
}

func TestResolveSingleHTMLFile(t *testing.T) {
	dir := realDir(t, t.TempDir())
	os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html></html>"), 0o644)
	rt, err := Resolve(filepath.Join(dir, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	if rt.Type != TypeWebsite || rt.EntryPoint != "index.html" || rt.RootDir != dir {
		t.Fatalf("unexpected resolution: %+v", rt)
	}
}

// realDir resolves any symlinks in dir (e.g. macOS's /var -> /private/var)
// so comparisons against Resolve's already-symlink-resolved output line up.
func realDir(t *testing.T, dir string) string {
	t.Helper()
	real, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	return real
}

func TestResolveDirectoryWithIndex(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html></html>"), 0o644)
	rt, err := Resolve(dir)
	if err != nil {
		t.Fatal(err)
	}
	if rt.Type != TypeWebsite {
		t.Fatalf("expected website type, got %s", rt.Type)
	}
}

func TestResolveDirectoryWithoutIndex(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("hi"), 0o644)
	rt, err := Resolve(dir)
	if err != nil {
		t.Fatal(err)
	}
	if rt.Type != TypeDirectory {
		t.Fatalf("expected directory type, got %s", rt.Type)
	}
}

func TestResolveNestedEntryPoint(t *testing.T) {
	dir := realDir(t, t.TempDir())
	os.MkdirAll(filepath.Join(dir, "dist"), 0o755)
	os.WriteFile(filepath.Join(dir, "dist", "index.html"), []byte("<html></html>"), 0o644)
	rt, err := Resolve(dir)
	if err != nil {
		t.Fatal(err)
	}
	if rt.Type != TypeWebsite || rt.RootDir != filepath.Join(dir, "dist") {
		t.Fatalf("expected to serve from dist/, got %+v", rt)
	}
}

func TestResolveMissingPath(t *testing.T) {
	if _, err := Resolve(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("expected error for missing path")
	}
}
