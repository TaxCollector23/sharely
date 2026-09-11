export function DevServer() {
  return (
    <section className="border-t border-[var(--border)] py-16">
      <div className="mx-auto max-w-5xl px-5 sm:px-6">
        <div className="grid gap-10 lg:grid-cols-[0.9fr_1.1fr] lg:items-center">
          <div>
            <h2 className="text-xl font-semibold tracking-tight sm:text-2xl">
              Already have a dev server running?
            </h2>
            <p className="mt-3 max-w-md text-[15px] leading-relaxed text-[var(--text-muted)]">
              Run{' '}
              <code className="rounded bg-[var(--code-bg)] px-1.5 py-0.5 font-mono text-[13px]">
                sharely start
              </code>{' '}
              in a directory where Vite (or anything similar) is already listening on{' '}
              <code className="rounded bg-[var(--code-bg)] px-1.5 py-0.5 font-mono text-[13px]">
                localhost:5173
              </code>
              , and Sharely proxies that server to the LAN instead of serving files itself. It
              never starts a dev server on its own — it only forwards to one you already have
              running.
            </p>
            <p className="mt-3 max-w-md text-sm leading-relaxed text-[var(--text-faint)]">
              WebSocket connections proxy through too, so hot reload keeps working on the phone or
              laptop you shared with — as long as the dev server itself allows proxied hosts.
            </p>
          </div>

          <div className="flex flex-col items-stretch gap-3 sm:flex-row sm:items-center sm:justify-center">
            <div className="rounded-md border border-[var(--border-strong)] px-4 py-3 text-center font-mono text-[13px]">
              localhost:5173
            </div>
            <div className="flex items-center justify-center gap-2 text-[var(--text-faint)] sm:flex-col">
              <span aria-hidden="true" className="text-lg leading-none sm:rotate-90">
                &rarr;
              </span>
              <span className="text-[11px] uppercase tracking-wide">Sharely</span>
            </div>
            <div className="rounded-md border border-[var(--accent)] bg-[var(--accent-soft)] px-4 py-3 text-center font-mono text-[13px] text-[var(--accent)]">
              sharely.local/bluebird
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
