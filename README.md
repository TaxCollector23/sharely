# Sharely

**The quickest way to move something between people and devices on the same network.**

Sharely turns any file, folder, static site, or local development server into a link that opens on nearby phones and computers. It runs directly from your machine: no upload, no cloud account, and no app required on the receiving device.

```bash
cd ~/Desktop/photos
sharely start
```

Sharely prints a link and a QR code, then continues running quietly in the background. The terminal is yours again immediately.

## Where Sharely fits

- Preview a local website on a real phone or tablet.
- Give someone a large folder without uploading it first.
- Move a file between devices that do not share an ecosystem.
- Show a work-in-progress to someone sitting beside you.
- Share a running Vite or other local dev server across the room.

The other person only needs a browser and access to the same Wi-Fi or LAN.

## Built around local sharing

- **Nothing leaves the network.** Sharely has no relay, storage service, account system, or telemetry.
- **The link is checked before it is shown.** If the friendly local hostname is unavailable, Sharely uses the machine's LAN address instead.
- **Shares expire.** The default lifetime is one hour; 15 minutes, four hours, and “until stopped” are also available.
- **Passwords are optional.** Add one when sharing on a busy or less-trusted network.
- **Sensitive files stay hidden.** Environment files, private keys, Git metadata, and common credential directories are excluded.
- **Local sites work as sites.** Relative CSS, scripts, images, and proxied WebSocket connections keep working.

## Install

Sharely currently supports macOS and Linux.

```bash
go install github.com/TaxCollector23/sharely/cmd/sharely@latest
```

An npm wrapper is also available; it builds the same Go binary during installation:

```bash
npm install -g sharely-cli
```

Prebuilt downloads and Homebrew installation are planned.

## Use Sharely

```bash
sharely start                  # share the current folder
sharely start report.pdf       # share one file
sharely start ./project        # share another folder
sharely start --password       # require a password
sharely start --expires 15m    # make a short-lived share
```

Once a share is running:

```bash
sharely list                   # see active shares
sharely qr <id>                # show its link and QR again
sharely logs <id>              # see recent requests
sharely dashboard              # open the local control center
sharely stop <id>              # stop one share
sharely stop-all               # stop every share
sharely doctor                 # diagnose network problems
```

Run `sharely help` for every option.

## A few honest limits

Sharely is deliberately local. It does not create public internet links, punch through routers, or replace a cloud drive. Devices must be on the same reachable network, and some guest or corporate Wi-Fi networks block device-to-device traffic.

The current installer requires Go. That is an installation constraint, not a runtime service dependency: Sharely itself is a single binary with its dashboard embedded.

## Open source

Sharely is MIT licensed. Issues and contributions are welcome. To work on it locally, clone the repository and run `go test ./...`; the marketing site and dashboard live under `web/`.

[View the source](https://github.com/TaxCollector23/sharely) · [Report a problem](https://github.com/TaxCollector23/sharely/issues)
