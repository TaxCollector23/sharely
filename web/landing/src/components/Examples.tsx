export function Examples() {
  return (
    <section className="border-t border-[var(--border)] py-16">
      <div className="mx-auto max-w-5xl px-5 sm:px-6">
        <h2 className="text-xl font-semibold tracking-tight sm:text-2xl">
          Works with whatever&rsquo;s already in the folder.
        </h2>
        <p className="mt-2 max-w-lg text-[15px] text-[var(--text-muted)]">
          Same command, three targets. Sharely looks at what you pointed it at and figures out how
          to serve it.
        </p>

        <div className="mt-10 space-y-6">
          {/* File: single compact row */}
          <div className="flex flex-col gap-3 rounded-lg border border-[var(--border)] p-5 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex items-center gap-3">
              <FileGlyph />
              <div>
                <p className="text-[15px] font-medium">A single file</p>
                <p className="font-mono text-[13px] text-[var(--text-muted)]">
                  $ sharely start report.pdf
                </p>
              </div>
            </div>
            <p className="font-mono text-[13px] text-[var(--accent)] sm:text-right">
              sharely.local/report-pdf
            </p>
          </div>

          {/* Folder: wider block with a small tree */}
          <div className="grid gap-6 rounded-lg border border-[var(--border)] p-5 sm:grid-cols-[1fr_auto] sm:items-center">
            <div>
              <p className="text-[15px] font-medium">A whole folder</p>
              <p className="mt-1 font-mono text-[13px] text-[var(--text-muted)]">
                $ sharely start ./project
              </p>
              <p className="mt-3 max-w-sm text-sm leading-relaxed text-[var(--text-muted)]">
                Sharely serves the directory as-is, so nested files and folders stay reachable at
                their normal paths.
              </p>
            </div>
            <div className="rounded border border-[var(--code-border)] bg-[var(--code-bg)] px-4 py-3 font-mono text-[12px] leading-[1.7] text-[var(--text-muted)] sm:min-w-[220px]">
              <p className="text-[var(--text)]">project/</p>
              <p>&nbsp;&nbsp;├─ index.html</p>
              <p>&nbsp;&nbsp;├─ styles.css</p>
              <p>&nbsp;&nbsp;└─ assets/</p>
              <p className="mt-2 text-[var(--accent)]">↳ sharely.local/project</p>
            </div>
          </div>

          {/* Site: mini browser chrome mockup */}
          <div className="grid gap-6 rounded-lg border border-[var(--border)] p-5 sm:grid-cols-[auto_1fr] sm:items-center">
            <div className="overflow-hidden rounded border border-[var(--border-strong)] sm:w-64">
              <div className="flex items-center gap-1.5 border-b border-[var(--border)] bg-[var(--bg-sunken)] px-2.5 py-1.5">
                <span className="h-1.5 w-1.5 rounded-full bg-[var(--border-strong)]" />
                <span className="h-1.5 w-1.5 rounded-full bg-[var(--border-strong)]" />
                <span className="ml-1 truncate rounded-sm bg-[var(--bg-raised)] px-2 py-0.5 font-mono text-[10px] text-[var(--text-faint)]">
                  sharely.local/bluebird
                </span>
              </div>
              <div className="space-y-1.5 bg-[var(--bg-raised)] px-3 py-4">
                <div className="h-2 w-3/4 rounded-sm bg-[var(--border)]" />
                <div className="h-2 w-1/2 rounded-sm bg-[var(--border)]" />
                <div className="h-8 w-full rounded-sm bg-[var(--accent-soft)]" />
              </div>
            </div>
            <div>
              <p className="text-[15px] font-medium">A static site</p>
              <p className="mt-1 font-mono text-[13px] text-[var(--text-muted)]">
                $ sharely start index.html
              </p>
              <p className="mt-3 max-w-sm text-sm leading-relaxed text-[var(--text-muted)]">
                Point Sharely at the entry file and it correctly serves relative assets — CSS, JS,
                images — from the same folder.
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}

function FileGlyph() {
  return (
    <svg width="30" height="30" viewBox="0 0 24 24" fill="none" aria-hidden="true" className="shrink-0">
      <path
        d="M6 2h8l4 4v16H6V2Z"
        stroke="var(--text-faint)"
        strokeWidth="1.4"
        strokeLinejoin="round"
      />
      <path d="M14 2v4h4" stroke="var(--text-faint)" strokeWidth="1.4" strokeLinejoin="round" />
    </svg>
  )
}
