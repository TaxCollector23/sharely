package server

import (
	"fmt"
	"net/http"
)

// pageCSS is shared by every Sharely-rendered page (auth, expired, error,
// file browser chrome) so the receiving-device experience feels coherent
// without ever looking like a "platform" wrapping the user's content.
const pageCSS = `
:root{color-scheme:light dark;--bg:#fafafa;--fg:#18181b;--muted:#71717a;--border:#e4e4e7;--accent:#2563eb;--card:#ffffff}
@media(prefers-color-scheme:dark){:root{--bg:#0b0b0c;--fg:#f4f4f5;--muted:#a1a1aa;--border:#2a2a2e;--card:#141416}}
*{box-sizing:border-box}
body.sharely-page{margin:0;min-height:100vh;display:flex;align-items:center;justify-content:center;background:var(--bg);color:var(--fg);
font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Inter,sans-serif;padding:24px}
.sharely-card{max-width:420px;width:100%;background:var(--card);border:1px solid var(--border);border-radius:10px;padding:32px}
.sharely-card h1{font-size:19px;margin:0 0 8px;font-weight:600}
.muted{color:var(--muted);font-size:14px;line-height:1.5;margin:0 0 20px}
.sharely-card form{display:flex;flex-direction:column;gap:10px}
.sharely-card input{border:1px solid var(--border);background:transparent;color:var(--fg);border-radius:6px;padding:10px 12px;font-size:14px}
.sharely-card input:focus{outline:2px solid var(--accent);outline-offset:1px}
.sharely-card button{border:none;background:var(--accent);color:#fff;border-radius:6px;padding:10px 12px;font-size:14px;font-weight:500;cursor:pointer}
.sharely-card button:hover{opacity:.92}
.sharely-card .err{color:#dc2626;font-size:13px;margin:-8px 0 12px}
.sharely-mark{font-size:12px;color:var(--muted);letter-spacing:.02em;margin-top:20px}
a.btn{display:inline-block;text-decoration:none;border:1px solid var(--border);border-radius:6px;padding:9px 14px;font-size:14px;color:var(--fg);margin-top:8px}
a.btn:hover{background:var(--border)}
`

func writeHTML(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprint(w, body)
}

func expiredPage() string {
	return `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>Share unavailable · Sharely</title><style>` + pageCSS + `</style></head>
<body class="sharely-page"><main class="sharely-card">
<h1>Share unavailable</h1>
<p class="muted">This share has ended. Ask the person who shared it to create a new link.</p>
<div class="sharely-mark">Sharely</div>
</main></body></html>`
}

func notFoundPage() string {
	return `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>Not found · Sharely</title><style>` + pageCSS + `</style></head>
<body class="sharely-page"><main class="sharely-card">
<h1>Nothing here</h1>
<p class="muted">There's no active share at this address.</p>
<div class="sharely-mark">Sharely</div>
</main></body></html>`
}

func passwordPage(shareID string, wrong bool) string {
	errMsg := ""
	if wrong {
		errMsg = `<p class="err">That password isn't right.</p>`
	}
	return `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>Password required · Sharely</title><style>` + pageCSS + `</style></head>
<body class="sharely-page"><main class="sharely-card">
<h1>This share is password-protected</h1>
<p class="muted">Enter the password to continue.</p>
` + errMsg + `
<form method="POST" action="/` + shareID + `/_auth">
<input type="password" name="password" autofocus autocomplete="current-password" placeholder="Password">
<button type="submit">Continue</button>
</form>
</main></body></html>`
}
