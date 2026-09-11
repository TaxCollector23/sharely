// Package dashboard embeds the built dashboard SPA into the sharely binary
// so a single compiled executable is fully self-contained — no Node runtime
// or separate asset directory required at install time. Run `npm run
// build` in this directory before `go build` to regenerate dist/.
package dashboard

import "embed"

//go:embed all:dist
var Dist embed.FS

// DistSubdir is the subdirectory within Dist that holds the site root,
// matching what embed.FS returns the files rooted at.
const DistSubdir = "dist"
