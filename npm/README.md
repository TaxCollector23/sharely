# sharely-cli

npm install wrapper for [Sharely](https://github.com/TaxCollector23/sharely) — share files, folders, and local sites over your network.

```bash
npm install -g sharely-cli
cd my-project
sharely start
```

This package builds the real `sharely` binary from source at install time using the Go toolchain (Go 1.22+ required for now — a prebuilt-binary release that doesn't need Go is planned).

> **If `sharely` fails right after install** saying the binary wasn't built: recent npm versions block a package's install script by default on some setups, and this package's install script is what actually compiles the binary. Reinstall allowing it explicitly:
>
> ```bash
> npm install -g --allow-scripts=sharely-cli sharely-cli
> ```

See the [main repository](https://github.com/TaxCollector23/sharely) for full documentation, or install directly with Go instead — this path never depends on npm's script policy:

```bash
go build -o sharely github.com/TaxCollector23/sharely/cmd/sharely
```

## License

MIT
