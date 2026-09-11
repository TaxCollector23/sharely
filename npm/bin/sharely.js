#!/usr/bin/env node
// Thin shim: forwards straight to the real Go binary built at install time
// by scripts/install.js. Kept intentionally tiny — all of sharely's actual
// behavior lives in the Go binary, not here.

const { spawnSync } = require('node:child_process')
const path = require('node:path')
const fs = require('node:fs')

const binPath = path.join(__dirname, 'sharely-bin')

if (!fs.existsSync(binPath)) {
  console.error(
    "sharely-cli: the sharely binary wasn't built.\n\n" +
      'Most likely cause: npm skipped this package\'s install script, which is\n' +
      'what actually builds sharely (npm 10+ blocks install scripts by default\n' +
      "on some setups). Reinstall allowing this package's script:\n\n" +
      '  npm install -g --allow-scripts=sharely-cli sharely-cli\n\n' +
      'or, if that flag is unavailable in your npm version:\n\n' +
      '  npm install -g --ignore-scripts=false sharely-cli\n\n' +
      'Otherwise, install with Go directly (no npm/script gate involved):\n\n' +
      '  go install github.com/TaxCollector23/sharely/cmd/sharely@latest\n',
  )
  process.exit(1)
}

const result = spawnSync(binPath, process.argv.slice(2), { stdio: 'inherit' })
process.exit(result.status ?? 1)
