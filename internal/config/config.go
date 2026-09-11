// Package config holds Sharely's runtime defaults and local paths.
package config

import (
	"os"
	"path/filepath"
)

const (
	// ControlHost is the loopback-only address for the dashboard/API server.
	ControlHost = "127.0.0.1"
	// ControlPortDefault is the first port tried for the dashboard/API server.
	ControlPortDefault = 47732
	// ContentPortDefault is the first port tried for the LAN content server.
	ContentPortDefault = 4821

	DefaultExpiry = "1h"
	LocalHostname = "sharely.local"

	// SessionCookiePrefix names the per-share auth cookie.
	SessionCookiePrefix = "sharely_auth_"
)

// Dir returns ~/.config/sharely, creating it if necessary. Sharely does not
// require this directory to run; it is only used for optional persisted
// preferences.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "sharely")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}
