const lines = [
  { cmd: 'sharely list', out: 'bluebird   ./project        expires in 42m' },
  { cmd: 'sharely stop bluebird', out: 'Stopped bluebird.' },
  { cmd: 'sharely doctor', out: 'Network OK · sharely.local resolves · port 4821 free' },
  { cmd: 'sharely logs bluebird', out: '127.0.0.1 GET /assets/app.js 200' },
]

export function Cli() {
  return (
    <section id="cli" className="border-t border-[var(--border)] py-16">
      <div className="mx-auto max-w-5xl px-5 sm:px-6">
        <div className="grid gap-10 lg:grid-cols-[0.85fr_1.15fr]">
          <div>
            <h2 className="text-xl font-semibold tracking-tight sm:text-2xl">
              A CLI, not a dashboard you have to babysit.
            </h2>
            <p className="mt-3 max-w-sm text-[15px] leading-relaxed text-[var(--text-muted)]">
              Check what&rsquo;s running, tail the request log, or shut everything down at once —
              all from the terminal you already have open.
            </p>
            <dl className="mt-6 space-y-3 text-sm">
              <div className="flex gap-3">
                <dt className="w-28 shrink-0 font-mono text-[var(--text-faint)]">--expires</dt>
                <dd className="text-[var(--text-muted)]">15m · 1h · 4h · until-stopped</dd>
              </div>
              <div className="flex gap-3">
                <dt className="w-28 shrink-0 font-mono text-[var(--text-faint)]">--password</dt>
                <dd className="text-[var(--text-muted)]">require a password to open the link</dd>
              </div>
              <div className="flex gap-3">
                <dt className="w-28 shrink-0 font-mono text-[var(--text-faint)]">--port</dt>
                <dd className="text-[var(--text-muted)]">pin the server to a specific port</dd>
              </div>
              <div className="flex gap-3">
                <dt className="w-28 shrink-0 font-mono text-[var(--text-faint)]">stop-all</dt>
                <dd className="text-[var(--text-muted)]">tear down every active share</dd>
              </div>
            </dl>
            <p className="mt-6 text-sm text-[var(--text-faint)]">
              Bare <code className="font-mono">sharely</code> is shorthand for{' '}
              <code className="font-mono">sharely start</code> — both share the current directory.
            </p>
          </div>

          <div className="rounded-lg border border-[var(--border-strong)] bg-[var(--terminal-bg)] px-5 py-5 font-mono text-[13px] leading-[1.9] text-[var(--terminal-text)]">
            {lines.map((line) => (
              <div key={line.cmd} className="mb-4 last:mb-0">
                <p>
                  <span className="text-[var(--terminal-accent)]">$</span> {line.cmd}
                </p>
                <p className="text-[var(--terminal-muted)]">{line.out}</p>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}
