// Package security implements the path-containment boundary that keeps a
// share from ever exposing anything outside its root directory.
package security

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ErrOutsideRoot is returned whenever a resolved path would escape the
// share's root directory, whether via "..", an absolute path, or a symlink.
var ErrOutsideRoot = errors.New("security: path escapes share root")

// sensitivePatterns are excluded from directory listings and direct access
// by default, even when they live inside the shared root, because they
// commonly hold secrets a user would not intend to publish.
var sensitivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(^|/)\.env(\..*)?$`),
	regexp.MustCompile(`(^|/)id_rsa$`),
	regexp.MustCompile(`(^|/)id_ed25519$`),
	regexp.MustCompile(`\.pem$`),
	regexp.MustCompile(`\.key$`),
	regexp.MustCompile(`(^|/)\.git(/|$)`),
	regexp.MustCompile(`(^|/)\.ssh(/|$)`),
	regexp.MustCompile(`(^|/)\.aws(/|$)`),
}

// IsSensitive reports whether relPath (slash-separated, relative to the
// share root) matches one of Sharely's default sensitive-file exclusions.
func IsSensitive(relPath string) bool {
	clean := strings.TrimPrefix(filepath.ToSlash(relPath), "/")
	for _, p := range sensitivePatterns {
		if p.MatchString(clean) {
			return true
		}
	}
	return false
}

// Resolve safely joins root and an untrusted, URL-decoded request path,
// guaranteeing the result stays within root even in the presence of "..",
// absolute-path segments, or a symlink chain that would otherwise escape.
// root itself must already be an absolute, symlink-resolved path.
func Resolve(root, requestPath string) (string, error) {
	// Normalize backslashes (Windows-style traversal attempts) and strip any
	// leading slash so filepath.Join treats it as relative.
	clean := strings.ReplaceAll(requestPath, "\\", "/")
	clean = strings.TrimPrefix(clean, "/")

	joined := filepath.Join(root, clean)
	joined = filepath.Clean(joined)

	if !withinRoot(root, joined) {
		return "", ErrOutsideRoot
	}

	// Resolve symlinks on the deepest existing ancestor so a symlink cannot
	// redirect us outside root. os.Lstat rather than Stat to catch a
	// symlink at the final component too.
	resolved, err := resolveExisting(joined)
	if err != nil {
		return "", err
	}
	if !withinRoot(root, resolved) {
		return "", ErrOutsideRoot
	}
	return joined, nil
}

func withinRoot(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// resolveExisting walks up from path until it finds the longest existing
// ancestor, fully resolves every symlink along that ancestor's path
// (filepath.EvalSymlinks handles symlinked components anywhere in the
// chain, not just the final one), then re-appends the remaining
// not-yet-existing suffix.
func resolveExisting(path string) (string, error) {
	suffix := ""
	cur := path
	for {
		if _, err := os.Stat(cur); err == nil {
			real, err := filepath.EvalSymlinks(cur)
			if err != nil {
				return "", err
			}
			return filepath.Join(real, suffix), nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return path, nil
		}
		suffix = filepath.Join(filepath.Base(cur), suffix)
		cur = parent
	}
}
