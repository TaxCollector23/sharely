# sharely-cli

npm install wrapper for [Sharely](https://github.com/TaxCollector23/sharely) — share files, folders, and local sites over your network.

```bash
npm install -g sharely-cli
cd my-project
sharely start
```

This package builds the real `sharely` binary from source at install time using the Go toolchain (Go 1.22+ required for now — a prebuilt-binary release that doesn't need Go is planned). See the [main repository](https://github.com/TaxCollector23/sharely) for full documentation, or install directly with Go:

```bash
go build -o sharely github.com/TaxCollector23/sharely/cmd/sharely
```

## License

MIT
