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
    "sharely-cli: the sharely binary wasn't built. Try reinstalling:\n\n" +
      '  npm install -g sharely-cli\n',
  )
  process.exit(1)
}

const result = spawnSync(binPath, process.argv.slice(2), { stdio: 'inherit' })
process.exit(result.status ?? 1)
