import { useState } from 'react'

const repo = 'https://github.com/TaxCollector23/sharely'
const installCommand = 'go install github.com/TaxCollector23/sharely/cmd/sharely@latest'

function CopyButton({ value, compact = false }: { value: string; compact?: boolean }) {
  const [copied, setCopied] = useState(false)
  async function copy() {
    await navigator.clipboard.writeText(value)
    setCopied(true)
    window.setTimeout(() => setCopied(false), 1600)
  }
  return <button className={compact ? 'copy copy--compact' : 'copy'} onClick={copy} type="button">{copied ? 'Copied!' : 'Copy'}</button>
}

function Logo() {
  return <span className="logo-mark" aria-hidden="true"><i /><i /><i /><svg viewBox="0 0 44 44"><path d="M14 22 28 13M14 22l14 9" /></svg></span>
}

function ArrowIcon() {
  return <svg viewBox="0 0 20 20" aria-hidden="true"><path d="M4 10h12M11 5l5 5-5 5" /></svg>
}

function App() {
  return (
    <div className="site-shell">
      <header className="nav wrap">
        <a href="#top" className="brand"><Logo />Sharely</a>
        <nav aria-label="Main navigation">
          <a href="#why">Why Sharely</a><a href="#how">How it works</a><a href="#safety">Privacy</a>
        </nav>
        <a className="nav-github" href={repo} target="_blank" rel="noreferrer">GitHub <span>↗</span></a>
      </header>

      <main id="top">
        <section className="hero wrap">
          <div className="hero-copy">
            <div className="eyebrow"><span /> Local file sharing, minus the ceremony</div>
            <h1>From your folder.<br /><em>To their screen.</em></h1>
            <p className="hero-lede">Share any file, folder, or local website with people on the same Wi-Fi. One command creates a link. Nothing gets uploaded.</p>
            <div className="hero-actions">
              <a className="button button--dark" href="#install">Install Sharely <ArrowIcon /></a>
              <a className="text-link" href="#how">See how it works <span>↓</span></a>
            </div>
            <div className="trust-row"><span>No account</span><span>No cloud</span><span>Open source</span></div>
          </div>

          <div className="network-demo" aria-label="A Sharely link connecting a laptop to nearby devices">
            <div className="orbit orbit--one" /><div className="orbit orbit--two" />
            <div className="device device--laptop"><div className="laptop-screen"><span className="screen-dot" /><span className="screen-line" /><span className="screen-line short" /></div><div className="laptop-base" /><b>Your laptop</b></div>
            <div className="device device--phone"><div className="phone-screen"><span>sharely</span><b>photos/</b><i>12 files</i></div><b>Alex’s phone</b></div>
            <div className="device device--tablet"><div className="tablet-screen"><span className="mini-card" /><span className="mini-card" /></div><b>iPad</b></div>
            <div className="share-pill"><span className="pulse" /> sharely.local/fig</div>
            <div className="signal signal--a" /><div className="signal signal--b" />
          </div>

          <div className="command-bar">
            <div className="command-label"><span>01</span> Open Terminal</div>
            <code><i>$</i> sharely start</code><CopyButton value="sharely start" compact />
            <div className="command-result"><span>02</span><b>Link ready</b><small>Share it or scan the QR</small></div>
          </div>
        </section>

        <section className="ticker" aria-label="Sharely can share"><div><span>FILES</span><i>✦</i><span>FOLDERS</span><i>✦</i><span>STATIC SITES</span><i>✦</i><span>DEV SERVERS</span><i>✦</i><span>FILES</span><i>✦</i><span>FOLDERS</span></div></section>

        <section className="why wrap" id="why">
          <div className="section-intro"><span className="section-number">01 / WHY SHARELY</span><h2>The fastest route between two nearby screens.</h2><p>When AirDrop is unavailable, email is absurd, and deploying is overkill.</p></div>
          <div className="use-grid">
            <article className="use-card use-card--wide">
              <span className="card-kicker">FOR MAKERS</span>
              <div className="browser-art"><div><i /><i /><i /></div><main><span>localhost</span><strong>Your work,<br />on their device.</strong></main></div>
              <h3>Preview on a real device</h3><p>Open your local site on a phone, tablet, or a teammate’s laptop without deploying a thing.</p>
            </article>
            <article className="use-card use-card--lime"><span className="card-kicker">FOR EVERYONE</span><div className="file-stack"><i>JPG</i><i>PDF</i><i>ZIP</i></div><h3>Hand off the heavy stuff</h3><p>Move photos, videos, exports, or entire folders at local-network speed.</p></article>
            <article className="use-card use-card--dark"><span className="card-kicker">NO FRICTION</span><div className="qr-art" aria-hidden="true"><span /><span /><span /><i /><i /><i /><i /><i /><i /><i /><i /><i /></div><h3>They just open a link</h3><p>No app to install. No account to create. A browser is all the other person needs.</p></article>
          </div>
        </section>

        <section className="how" id="how"><div className="wrap">
          <div className="section-intro section-intro--row"><div><span className="section-number">02 / HOW IT WORKS</span><h2>Share in three beats.</h2></div><p>Sharely figures out whether you gave it a file, folder, site, or running dev server.</p></div>
          <ol className="steps">
            <li><span>1</span><div><b>Point</b><p>Open a terminal inside the folder you want to share.</p><code>cd ~/Desktop/photos</code></div></li>
            <li><span>2</span><div><b>Start</b><p>Sharely starts quietly in the background and gives you a verified local link.</p><code>sharely start</code></div></li>
            <li><span>3</span><div><b>Pass it on</b><p>Send the link or let someone scan the QR code. Stop it whenever you’re done.</p><code>sharely stop fig</code></div></li>
          </ol>
        </div></section>

        <section className="privacy wrap" id="safety">
          <div className="privacy-copy"><span className="section-number">03 / PRIVATE BY DESIGN</span><h2>Your files never take the scenic route.</h2><p>Sharely serves straight from your machine to devices on your local network. There’s no Sharely cloud, relay, upload queue, or tracking pixel in the middle.</p><a href={repo} target="_blank" rel="noreferrer" className="text-link">Inspect the source <span>↗</span></a></div>
          <div className="privacy-list">
            <div><span>01</span><b>Local network only</b><p>Nothing is exposed to the public internet.</p></div>
            <div><span>02</span><b>Automatic expiry</b><p>Choose 15 minutes, 1 hour, 4 hours, or until stopped.</p></div>
            <div><span>03</span><b>Protected when needed</b><p>Add a password for busy or less-trusted networks.</p></div>
            <div><span>04</span><b>Sensitive files hidden</b><p>Keys, environment files, and Git metadata stay out of view.</p></div>
          </div>
        </section>

        <section className="install wrap" id="install"><div className="install-card">
          <span className="section-number">READY WHEN YOU ARE</span><h2>Keep your files close.<br />Share them faster.</h2><p>Sharely is free, open source, and built for macOS and Linux.</p>
          <div className="install-command"><code>{installCommand}</code><CopyButton value={installCommand} /></div><small>Requires Go for installation today. Prebuilt downloads are coming.</small>
        </div></section>
      </main>

      <footer className="footer wrap"><a href="#top" className="brand"><Logo />Sharely</a><p>Made for the people on your Wi-Fi.</p><div><a href={repo} target="_blank" rel="noreferrer">GitHub ↗</a><a href={`${repo}/blob/master/LICENSE`} target="_blank" rel="noreferrer">MIT License</a></div></footer>
    </div>
  )
}

export default App
