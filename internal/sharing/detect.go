package sharing

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

// commonDevPorts are the ports Sharely checks for an already-running local
// dev server when the target looks like a JS/TS project. Sharely never
// starts one itself — it only offers to proxy an existing one.
var commonDevPorts = []int{5173, 3000, 8080, 4321, 5174, 4173, 8000, 1234}

// ResolvedTarget describes what Sharely decided to do with a CLI argument.
type ResolvedTarget struct {
	Type          Type
	RootDir       string
	EntryPoint    string
	ProxyUpstream string
	Warning       string
}

// Resolve inspects a filesystem path (or empty string for the current
// directory) and decides what kind of share it should become. It never
// executes anything in the target — only reads the filesystem and, for
// proxy detection, probes localhost ports that may already be listening.
func Resolve(arg string) (*ResolvedTarget, error) {
	path := arg
	if path == "" {
		path = "."
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, fmt.Errorf("%s does not exist", path)
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return nil, fmt.Errorf("%s does not exist", path)
	}

	if !info.IsDir() {
		return resolveFile(resolved)
	}
	return resolveDir(resolved)
}

func resolveFile(path string) (*ResolvedTarget, error) {
	dir := filepath.Dir(path)
	name := filepath.Base(path)
	ext := filepath.Ext(name)
	if ext == ".html" || ext == ".htm" {
		return &ResolvedTarget{Type: TypeWebsite, RootDir: dir, EntryPoint: name}, nil
	}
	// Any other single file (PDF, image, text, binary, ...) is shared as
	// itself: its "root" is its parent directory so path security applies
	// uniformly, but the entry point pins the browser straight to the file.
	return &ResolvedTarget{Type: TypeFile, RootDir: dir, EntryPoint: name}, nil
}

var siteEntryCandidates = []string{
	"index.html",
	filepath.Join("public", "index.html"),
	filepath.Join("dist", "index.html"),
	filepath.Join("build", "index.html"),
}

func resolveDir(dir string) (*ResolvedTarget, error) {
	// An already-running dev server takes priority: if this looks like a JS
	// project and a common dev port is alive, proxy it rather than serving
	// stale build output.
	if hasFile(dir, "package.json") {
		if port, ok := detectRunningDevServer(); ok {
			return &ResolvedTarget{
				Type:          TypeProxy,
				RootDir:       dir,
				ProxyUpstream: fmt.Sprintf("http://127.0.0.1:%d", port),
			}, nil
		}
	}

	for _, candidate := range siteEntryCandidates {
		full := filepath.Join(dir, candidate)
		if fileExists(full) {
			root := dir
			entry := candidate
			if filepath.Dir(candidate) != "." {
				// Serve from the subdirectory that actually contains the
				// site so relative asset paths resolve correctly.
				root = filepath.Join(dir, filepath.Dir(candidate))
				entry = filepath.Base(candidate)
			}
			return &ResolvedTarget{Type: TypeWebsite, RootDir: root, EntryPoint: entry}, nil
		}
	}

	return &ResolvedTarget{Type: TypeDirectory, RootDir: dir}, nil
}

func hasFile(dir, name string) bool {
	return fileExists(filepath.Join(dir, name))
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func detectRunningDevServer() (int, bool) {
	for _, port := range commonDevPorts {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 150*time.Millisecond)
		if err == nil {
			conn.Close()
			return port, true
		}
	}
	return 0, false
}
