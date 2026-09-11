const steps = [
  { label: 'Install', code: 'brew install sharely' },
  { label: 'Run', code: 'sharely start' },
  { label: 'Share', code: 'sharely.local/bluebird' },
]

export function Install() {
  return (
    <section id="install" className="border-t border-[var(--border)] py-16">
      <div className="mx-auto max-w-5xl px-5 sm:px-6">
        <h2 className="text-xl font-semibold tracking-tight sm:text-2xl">Get started</h2>
        <p className="mt-2 max-w-md text-[15px] text-[var(--text-muted)]">
          One command to install, one to run.
        </p>

        <div className="mt-8 flex flex-col gap-0 sm:flex-row sm:items-stretch">
          {steps.map((step, i) => (
            <div key={step.label} className="flex flex-1 items-center">
              <div className="flex-1 rounded-lg border border-[var(--border-strong)] p-5">
                <p className="text-[11px] font-medium uppercase tracking-wide text-[var(--text-faint)]">
                  {step.label}
                </p>
                <p className="mt-2 font-mono text-[14px] text-[var(--text)]">{step.code}</p>
              </div>
              {i < steps.length - 1 && (
                <span
                  aria-hidden="true"
                  className="mx-3 hidden shrink-0 text-lg text-[var(--text-faint)] sm:block"
                >
                  &rarr;
                </span>
              )}
              {i < steps.length - 1 && (
                <span aria-hidden="true" className="my-2 block text-center text-lg text-[var(--text-faint)] sm:hidden">
                  &darr;
                </span>
              )}
            </div>
          ))}
        </div>

        <p className="mt-6 text-sm text-[var(--text-faint)]">
          Sharely is a single compiled binary — the dashboard is embedded in it, so there&rsquo;s
          nothing else to install. The Homebrew formula isn&rsquo;t published yet, so for now build
          it from source with <code className="font-mono">go build</code>.
        </p>
      </div>
    </section>
  )
}
