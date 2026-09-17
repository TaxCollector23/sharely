package sharing

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveProjectPrefersBuiltSite(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"scripts":{"build":"vite build"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<script type="module" src="/src/main.tsx"></script>`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "dist"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "dist", "index.html"), []byte(`<script src="/assets/app.js"></script>`), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve(dir)
	if err != nil {
		t.Fatal(err)
	}
	resolvedDir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != TypeWebsite || got.RootDir != filepath.Join(resolvedDir, "dist") || got.EntryPoint != "index.html" {
		t.Fatalf("expected built site, got %#v", got)
	}
}
