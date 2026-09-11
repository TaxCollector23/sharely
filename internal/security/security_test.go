package security

import (
	"os"
	"path/filepath"
	"testing"
)

func setupRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "sub", "nested.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	resolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

func TestResolveAllowsWithinRoot(t *testing.T) {
	root := setupRoot(t)
	got, err := Resolve(root, "file.txt")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if got != filepath.Join(root, "file.txt") {
		t.Fatalf("unexpected resolved path: %s", got)
	}

	if _, err := Resolve(root, "sub/nested.txt"); err != nil {
		t.Fatalf("expected success for nested file, got %v", err)
	}
}

func TestResolveBlocksTraversal(t *testing.T) {
	root := setupRoot(t)
	cases := []string{
		"../../../etc/passwd",
		"../etc/passwd",
		"sub/../../etc/passwd",
		"..%2F..%2Fetc%2Fpasswd", // not decoded here, but must not escape if handler decodes first and re-checks
		"/etc/passwd",
		"....//....//etc/passwd",
	}
	for _, c := range cases {
		got, err := Resolve(root, c)
		if err == nil && !withinRoot(root, got) {
			t.Errorf("case %q escaped root: %s", c, got)
		}
	}
}

func TestResolveBlocksSymlinkEscape(t *testing.T) {
	root := setupRoot(t)
	outside := t.TempDir()
	secret := filepath.Join(outside, "secret.txt")
	if err := os.WriteFile(secret, []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unsupported: %v", err)
	}

	_, err := Resolve(root, "escape/secret.txt")
	if err != ErrOutsideRoot {
		t.Fatalf("expected ErrOutsideRoot for symlink escape, got %v", err)
	}
}

func TestIsSensitive(t *testing.T) {
	sensitive := []string{".env", ".env.local", "id_rsa", "server.pem", "a.key", ".git/config", ".ssh/id_rsa"}
	for _, p := range sensitive {
		if !IsSensitive(p) {
			t.Errorf("expected %q to be flagged sensitive", p)
		}
	}
	safe := []string{"index.html", "notes.md", "envelope.txt", "photo.png"}
	for _, p := range safe {
		if IsSensitive(p) {
			t.Errorf("did not expect %q to be flagged sensitive", p)
		}
	}
}
