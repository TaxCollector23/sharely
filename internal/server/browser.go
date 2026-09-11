package server

import (
	"fmt"
	"html"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"

	"sharely/internal/mimekind"
)

type entryRow struct {
	Name  string
	Href  string
	IsDir bool
	Size  string
	Kind  string
}

func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for n2 := n / unit; n2 >= unit; n2 /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(n)/float64(div), units[exp])
}

func iconFor(kind string, isDir bool) string {
	if isDir {
		return "📁"
	}
	switch mimekind.Category(kind) {
	case mimekind.CategoryHTML:
		return "🌐"
	case mimekind.CategoryPDF:
		return "📄"
	case mimekind.CategoryImage:
		return "🖼"
	case mimekind.CategoryAudio:
		return "🎵"
	case mimekind.CategoryVideo:
		return "🎬"
	case mimekind.CategoryText:
		return "📝"
	default:
		return "📦"
	}
}

// renderBrowser lists the contents of dirPath (an absolute, already
// security-checked directory) for urlBase (the share-relative URL prefix,
// e.g. "/bluebird/src").
func renderBrowser(shareID, shareName, urlBase, dirPath string) (string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", err
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})

	var rows []entryRow
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		href := path.Join(urlBase, url.PathEscape(e.Name()))
		if e.IsDir() {
			href += "/"
		}
		_, category := mimekind.Detect(e.Name())
		size := ""
		if !e.IsDir() {
			size = humanSize(info.Size())
		}
		rows = append(rows, entryRow{
			Name:  e.Name(),
			Href:  href,
			IsDir: e.IsDir(),
			Size:  size,
			Kind:  string(category),
		})
	}

	crumbs := breadcrumbs(shareID, urlBase)

	var b strings.Builder
	b.WriteString(`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">`)
	b.WriteString(`<title>` + html.EscapeString(shareName) + ` · Sharely</title><style>` + browserCSS + `</style></head><body>`)
	b.WriteString(`<header class="sh-header"><div class="sh-name">` + html.EscapeString(shareName) + `</div><nav class="sh-crumbs">` + crumbs + `</nav></header>`)
	b.WriteString(`<main class="sh-list">`)
	if len(rows) == 0 {
		b.WriteString(`<div class="sh-empty">This folder is empty.</div>`)
	}
	for _, row := range rows {
		kindLabel := row.Size
		if row.IsDir {
			kindLabel = "Folder"
		}
		b.WriteString(`<a class="sh-row" href="` + html.EscapeString(row.Href) + `">`)
		b.WriteString(`<span class="sh-icon">` + iconFor(row.Kind, row.IsDir) + `</span>`)
		b.WriteString(`<span class="sh-fname">` + html.EscapeString(row.Name) + `</span>`)
		b.WriteString(`<span class="sh-fsize">` + html.EscapeString(kindLabel) + `</span>`)
		b.WriteString(`</a>`)
	}
	b.WriteString(`</main><footer class="sh-foot">Shared with Sharely</footer></body></html>`)
	return b.String(), nil
}

func breadcrumbs(shareID, urlBase string) string {
	trimmed := strings.Trim(strings.TrimPrefix(urlBase, "/"+shareID), "/")
	parts := []string{}
	if trimmed != "" {
		parts = strings.Split(trimmed, "/")
	}
	var b strings.Builder
	b.WriteString(`<a href="/` + shareID + `/">` + html.EscapeString(shareID) + `</a>`)
	cur := "/" + shareID
	for _, p := range parts {
		cur += "/" + url.PathEscape(p)
		b.WriteString(` <span class="sh-sep">/</span> <a href="` + cur + `/">` + html.EscapeString(p) + `</a>`)
	}
	return b.String()
}

const browserCSS = `
:root{color-scheme:light dark;--bg:#fafafa;--fg:#18181b;--muted:#71717a;--border:#e4e4e7;--card:#ffffff}
@media(prefers-color-scheme:dark){:root{--bg:#0b0b0c;--fg:#f4f4f5;--muted:#a1a1aa;--border:#2a2a2e;--card:#141416}}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--fg);font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Inter,sans-serif}
.sh-header{padding:20px 24px;border-bottom:1px solid var(--border);position:sticky;top:0;background:var(--bg)}
.sh-name{font-weight:600;font-size:15px;margin-bottom:6px}
.sh-crumbs{font-size:13px;color:var(--muted)}
.sh-crumbs a{color:var(--muted);text-decoration:none}
.sh-crumbs a:hover{color:var(--fg);text-decoration:underline}
.sh-sep{margin:0 2px}
.sh-list{max-width:720px;margin:0 auto;padding:8px 24px 60px}
.sh-row{display:flex;align-items:center;gap:12px;padding:11px 4px;text-decoration:none;color:var(--fg);border-bottom:1px solid var(--border)}
.sh-row:hover{background:var(--card)}
.sh-row:focus-visible{outline:2px solid #2563eb;outline-offset:-2px}
.sh-icon{font-size:16px;width:20px;text-align:center}
.sh-fname{flex:1;font-size:14px;overflow-wrap:anywhere}
.sh-fsize{font-size:12px;color:var(--muted);font-variant-numeric:tabular-nums}
.sh-empty{padding:40px 4px;color:var(--muted);font-size:14px;text-align:center}
.sh-foot{text-align:center;color:var(--muted);font-size:12px;padding:20px}
`
