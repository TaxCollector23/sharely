# sharely-cli

Share files, folders, and local websites with people on the same Wi-Fi.

```bash
npm install -g sharely-cli
cd my-project
sharely start
```

Sharely prints a link and a QR code. The other person opens it in a browser—no account or app required. Shares stay on your local network and expire automatically.

This package builds the Sharely binary from source during installation, so Go 1.27+ is required for now. If your npm configuration blocks install scripts, allow this package's script and reinstall:

```bash
npm install -g --allow-scripts=sharely-cli sharely-cli
```

See the [main repository](https://github.com/TaxCollector23/sharely) for commands, privacy details, and support.

MIT License
