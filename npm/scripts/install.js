#!/usr/bin/env node
// Builds the real sharely binary from source at install time.
//
// There's no prebuilt-binary release pipeline yet, so this fetches the
// tagged source matching this package's version (falling back to the
// default branch if that tag doesn't exist yet) and compiles it locally
// with the Go toolchain. This is a deliberate, temporary approach: it's
// honest about what it does, and it fails loudly with clear instructions
// rather than silently leaving behind a broken `sharely` command.

const { spawnSync } = require('node:child_process')
const fs = require('node:fs')
const os = require('node:os')
const path = require('node:path')

const REPO = 'TaxCollector23/sharely'
const pkg = require('../package.json')
const tag = 'v' + pkg.version

function run(cmd, args, opts = {}) {
  return spawnSync(cmd, args, { stdio: 'inherit', ...opts })
}

function have(cmd) {
  const res = spawnSync(cmd, ['version'], { stdio: 'ignore' })
  return !res.error
}

function fail(message) {
  console.error('\n' + message + '\n')
  process.exit(1)
}

if (process.platform !== 'darwin' && process.platform !== 'linux') {
  fail(
    'sharely-cli currently only supports macOS and Linux.\n' +
      'Windows support is planned — see https://github.com/' + REPO,
  )
}

if (!have('go')) {
  fail(
    'sharely-cli builds the real sharely binary from source, and that needs Go 1.22+.\n\n' +
      '  Install Go:  https://go.dev/dl/\n' +
      '  Then run:    npm install -g sharely-cli\n\n' +
      "(A prebuilt-binary release that doesn't need Go is planned.)",
  )
}
if (!have('curl') || !have('tar')) {
  fail('sharely-cli needs `curl` and `tar` on PATH to fetch and build its source.')
}

const binDir = path.join(__dirname, '..', 'bin')
fs.mkdirSync(binDir, { recursive: true })

const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'sharely-src-'))
const tarPath = path.join(tmpDir, 'src.tar.gz')

function downloadRef(ref) {
  const url = `https://codeload.github.com/${REPO}/tar.gz/${ref}`
  const res = run('curl', ['-fsSL', url, '-o', tarPath])
  return res.status === 0
}

console.log(`sharely-cli: fetching source (${tag})…`)
if (!downloadRef(`refs/tags/${tag}`)) {
  console.log(`sharely-cli: tag ${tag} not found yet, using the default branch instead…`)
  if (!downloadRef('HEAD')) {
    fail(`Couldn't download source from https://github.com/${REPO}. Check your network connection.`)
  }
}

const extracted = run('tar', ['-xzf', tarPath, '-C', tmpDir, '--strip-components=1'])
if (extracted.status !== 0) {
  fail('Failed to extract the downloaded source archive.')
}

console.log('sharely-cli: building with Go (this only happens once)…')
const built = run('go', ['build', '-o', path.join(binDir, 'sharely-bin'), './cmd/sharely'], {
  cwd: tmpDir,
})
if (built.status !== 0) {
  fail('`go build` failed — see the output above.')
}

fs.chmodSync(path.join(binDir, 'sharely-bin'), 0o755)
fs.rmSync(tmpDir, { recursive: true, force: true })

console.log('sharely-cli: built successfully. Run `sharely start` in any folder to try it.')
