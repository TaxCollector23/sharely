import { useState } from 'react'

const installOptions = [
  { label: 'Go', code: 'go install github.com/TaxCollector23/sharely/cmd/sharely@latest' },
  { label: 'npm', code: 'npm install -g sharely-cli' },
]

const nextSteps = [
  { label: 'Run', code: 'sharely start' },
  { label: 'Share', code: 'sharely.local/bluebird' },
]

function CopyableCommand({ code }: { code: string }) {
  const [copied, setCopied] = useState(false)

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(code)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {
      // Clipboard API unavailable — the command is still visible to copy by hand.
    }
  }

  return (
    <div className="flex items-center gap-2 rounded-lg border border-[var(--border-strong)] bg-[var(--code-bg)] px-4 py-3">
      <code className="min-w-0 flex-1 overflow-x-auto whitespace-nowrap font-mono text-[13px] text-[var(--text)] sm:text-sm">
        {code}
      </code>
      <button
        type="button"
        onClick={copy}
        className="shrink-0 rounded-md border border-[var(--border-strong)] px-2.5 py-1.5 text-xs font-medium text-[var(--text-muted)] transition-colors hover:border-[var(--text-faint)] hover:text-[var(--text)]"
        aria-label={`Copy command: ${code}`}
      >
        {copied ? 'Copied' : 'Copy'}
      </button>
    </div>
  )
}

export function Install() {
  return (
    <section id="install" className="border-t border-[var(--border)] py-16">
      <div className="mx-auto max-w-5xl px-5 sm:px-6">
        <h2 className="text-xl font-semibold tracking-tight sm:text-2xl">Get started</h2>
        <p className="mt-2 max-w-md text-[15px] text-[var(--text-muted)]">
          Two ways to install. Both build the real binary — there&rsquo;s no separate download
          server involved.
        </p>

        <div className="mt-6 grid gap-3 sm:grid-cols-2">
          {installOptions.map((opt) => (
            <div key={opt.label}>
              <p className="mb-1.5 text-[11px] font-medium uppercase tracking-wide text-[var(--text-faint)]">
                {opt.label}
              </p>
              <CopyableCommand code={opt.code} />
            </div>
          ))}
        </div>

        <div className="mt-8 flex flex-col gap-0 sm:flex-row sm:items-stretch">
          {nextSteps.map((step, i) => (
            <div key={step.label} className="flex flex-1 items-center">
              <div className="flex-1 rounded-lg border border-[var(--border-strong)] p-5">
                <p className="text-[11px] font-medium uppercase tracking-wide text-[var(--text-faint)]">
                  {step.label}
                </p>
                <p className="mt-2 font-mono text-[14px] text-[var(--text)]">{step.code}</p>
              </div>
              {i < nextSteps.length - 1 && (
                <span
                  aria-hidden="true"
                  className="mx-3 hidden shrink-0 text-lg text-[var(--text-faint)] sm:block"
                >
                  &rarr;
                </span>
              )}
              {i < nextSteps.length - 1 && (
                <span aria-hidden="true" className="my-2 block text-center text-lg text-[var(--text-faint)] sm:hidden">
                  &darr;
                </span>
              )}
            </div>
          ))}
        </div>

        <p className="mt-6 text-sm text-[var(--text-faint)]">
          Sharely is a single compiled binary — the dashboard is embedded in it via{' '}
          <code className="font-mono">go:embed</code>, so there&rsquo;s nothing else to install.
          Both install methods need Go 1.22+ on your machine for now; a prebuilt-binary release
          (Homebrew included) that skips that requirement is planned.
        </p>
      </div>
    </section>
  )
}
