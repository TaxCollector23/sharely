const steps = [
  {
    number: '01',
    title: 'Run sharely start',
    body: 'From inside a folder, on a single file, or with nothing running yet, Sharely figures out what you mean. If a dev server is already listening in that directory, it proxies that instead.',
    code: '$ sharely start',
  },
  {
    number: '02',
    title: 'It checks the network before handing you a link',
    body: 'Sharely verifies sharely.local actually resolves on this machine before using it. If it doesn’t, you get your real LAN IP instead — never a friendly-looking link that fails on the other device.',
    code: '→ sharely.local/bluebird',
  },
  {
    number: '03',
    title: 'Send the link, or scan the QR',
    body: 'Anyone on the same Wi-Fi opens it in a browser, no app required. The share expires on its own — default 1 hour — so there’s nothing to remember to clean up.',
    code: 'Available for 1 hour',
  },
]

export function HowItWorks() {
  return (
    <section id="how-it-works" className="border-t border-[var(--border)] py-16">
      <div className="mx-auto max-w-5xl px-5 sm:px-6">
        <h2 className="text-xl font-semibold tracking-tight sm:text-2xl">How it works</h2>
        <div className="mt-10 grid gap-10 sm:grid-cols-3 sm:gap-8">
          {steps.map((step, i) => (
            <div key={step.number} className={i > 0 ? 'sm:border-l sm:border-[var(--border)] sm:pl-8' : ''}>
              <span className="font-mono text-sm text-[var(--text-faint)]">{step.number}</span>
              <h3 className="mt-2 text-[15px] font-semibold">{step.title}</h3>
              <p className="mt-2 text-sm leading-relaxed text-[var(--text-muted)]">{step.body}</p>
              <p className="mt-4 rounded border border-[var(--code-border)] bg-[var(--code-bg)] px-3 py-2 font-mono text-[12.5px] text-[var(--text)]">
                {step.code}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
