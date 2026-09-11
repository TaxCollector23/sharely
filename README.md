# Sharely

Share what's on your machine.

Give Sharely a file, folder, or local site. It puts it on your network and gives you a link — no account, no cloud, no upload, no configuration.

```bash
cd my-project
sharely start
```

Run that in a directory and Sharely shares it immediately (bare `sharely` does the same thing — `start` is just the explicit, easy-to-remember spelling):

```
Sharely

Sharing  my-project

✓ Your site is ready.

  sharely.local/bluebird

  Available for 1 hour

  Scan to open
```

## Install

With Go 1.22+:

```bash
go install github.com/TaxCollector23/sharely/cmd/sharely@latest
```

Or with npm (still builds the real binary from source via Go under the hood — see [`npm/`](npm/)):

```bash
npm install -g sharely-cli
```

Or from a clone:

```bash
go build -o sharely ./cmd/sharely
```

Either way you get a single self-contained executable — the dashboard is compiled directly into the binary via `go:embed`, so there's no separate asset directory to ship or lose track of.

(A Homebrew formula and prebuilt binaries for a Go-free install are planned.)

## Quick start

```bash
sharely start                # share the current directory
sharely start index.html     # share a website (relative assets just work)
sharely start ./project      # share a folder
sharely start report.pdf     # share a single file
```

`sharely start` always works from wherever you `cd` to. The first invocation on a machine starts Sharely's background daemon; running `sharely start` again from a different folder (even in a different terminal) just adds another independent share to it.

### Sharely never hands you a broken link

Before printing "Your site is ready," Sharely actually fetches the link it's about to show you — not just checks that a socket opens. If your LAN address turns out not to be reachable (a strict firewall, an unusual network setup, certain VPN/container configurations), it automatically checks a guaranteed-working `http://127.0.0.1:.../` link on the same machine and offers that instead, along with a note to run `sharely doctor`. The same is true of the friendly `sharely.local` hostname: Sharely only ever shows it as the primary link after confirming this machine can actually resolve it, falling back to your real LAN IP otherwise. You should never end up staring at a link that looks right and just doesn't load.

The first `sharely start` invocation on a machine starts Sharely's local daemon (a LAN-facing content server plus a loopback-only dashboard/API) and creates the first share. Every subsequent `sharely start <path>` from any terminal on the same machine talks to that daemon and adds another independent share — no need to keep the original terminal open.

## Commands

```
sharely start [path]     Share a file, folder, or site (defaults to the current directory)
sharely [path]           Same thing — "start" is just the explicit spelling
sharely list             List active shares
sharely stop <id>        Stop one share
sharely stop-all         Stop every share
sharely logs <id>        Show recent requests for a share
sharely doctor           Run network/port/hostname diagnostics
sharely help             Show usage
sharely version          Show the version
```

Options for `sharely start [path]`:

```
--port <n>       Use a specific port for the content server
--host <ip>      Bind to a specific address
--password       Require a password to view this share
--expires <t>    15m, 1h, 4h, or until-stopped (default: 1h)
--open           Open the link in your browser
--quiet          Print only the link
--verbose        Show technical logs
```

## What Sharely understands

- **A single file** (`sharely report.pdf`, `sharely photo.png`) is shared as itself; the browser's native viewer handles PDFs, images, audio, and video.
- **An HTML file** (`sharely index.html`) is served from its parent directory so relative assets (`style.css`, `app.js`, `images/`) resolve correctly.
- **A folder with a detected entry point** (`index.html`, `public/index.html`, `dist/index.html`, `build/index.html`) is served as a website.
- **A plain folder** gets a clean, minimal file browser — no cloud-drive clone, just names, sizes, and download links.
- **A folder with a `package.json` and an already-running dev server** (Vite, webpack-dev-server, etc. on a common port) is proxied to the LAN, including WebSocket upgrades for HMR where the framework supports it. Sharely never starts a dev server itself — it only proxies one that's already running.

## Every share expires

Shares default to **1 hour** and can be set to 15 minutes, 4 hours, or "Until I stop it." Changing the expiration never changes the URL. Stopping a share (`sharely stop <id>`, or Stop in the dashboard) is immediate and the link stops working right away.

## Security model

- **Path containment.** Every request is resolved and canonicalized against the share's root before serving; `../`, encoded traversal, absolute paths, and symlink escapes are all rejected. See `internal/security` and its test suite.
- **Sensitive files excluded by default.** `.env`, `.env.*`, `.git/`, `.ssh/`, `.aws/`, `id_rsa`, `*.pem`, and `*.key` are hidden from directory listings and direct access even inside a shared root.
- **No execution, ever.** Sharely serves files. It never runs shell scripts, package scripts, or binaries in a shared directory. Dev-server proxying only forwards HTTP/WebSocket traffic to a server the user already started themselves.
- **Origin isolation.** The LAN-facing content server (where untrusted, user-shared HTML lives) and the loopback-only dashboard/API server are separate listeners on separate ports. A malicious shared page cannot reach the control API — it isn't reachable from the LAN at all.
- **Optional password protection**, hashed with bcrypt, backed by a signed session cookie scoped to that one share.
- **Local network only.** No public tunnels, no port forwarding, no cloud relay, no telemetry, no accounts.

## Architecture

```
cmd/sharely/            CLI entry point, daemon/client wiring
internal/server/        LAN content server + loopback dashboard/API + auth + file browser
internal/sharing/       Share domain object, lifecycle, target detection
internal/security/      Path containment / traversal / symlink protection
internal/proxy/         Reverse proxy with WebSocket support for dev servers
internal/network/       LAN interface + free port discovery
internal/discovery/     Best-effort "sharely.local" mDNS responder
internal/qr/            Real QR code generation (SVG + terminal ANSI)
internal/mimekind/      Content-type / preview classification
web/dashboard/          React + Vite + Tailwind + shadcn/ui control center (served by the daemon)
web/landing/            Marketing site (deployed separately)
```

The first `sharely` invocation on a machine becomes the daemon; later invocations detect it (via the loopback API) and act as clients that register another share and exit, so multiple independent shares can run from one process without keeping every terminal open.

## Development

```bash
go build ./...
go test ./...
go run ./cmd/sharely
```

Dashboard:

```bash
cd web/dashboard
npm install
npm run dev      # local development against a running `sharely` daemon
npm run build    # outputs to dist/, which is go:embed'd into the sharely binary
```

`dist/` is committed to the repository (see `web/dashboard/embed.go`) specifically so a fresh clone can run `go build` immediately without a Node toolchain. After changing the dashboard, run `npm run build` and commit the updated `dist/` alongside your source changes. For local iteration without rebuilding, point the daemon at the on-disk build instead of the embedded copy:

```bash
SHARELY_DASHBOARD_DIR=web/dashboard/dist go run ./cmd/sharely
```

Landing page:

```bash
cd web/landing
npm install
npm run dev
```

## Platform support

macOS and Linux today. The OS-specific pieces (network interface detection, mDNS) are isolated in `internal/network` and `internal/discovery` so Windows support can be added later without touching the rest of the codebase.

## License

MIT
