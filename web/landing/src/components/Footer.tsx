export function Footer() {
  return (
    <footer className="border-t border-[var(--border)] py-10">
      <div className="mx-auto flex max-w-5xl flex-col gap-4 px-5 text-sm text-[var(--text-faint)] sm:flex-row sm:items-center sm:justify-between sm:px-6">
        <div>
          <p className="font-medium text-[var(--text-muted)]">Sharely</p>
          <p className="mt-1 max-w-sm">
            <code className="font-mono">sharely start</code> puts a file, folder, or site on your
            network. Local only — no cloud, no accounts.
          </p>
        </div>
        <div className="flex items-center gap-4">
          <a href="#" className="transition-colors hover:text-[var(--text)]">
            GitHub
          </a>
          <span>License: MIT</span>
        </div>
      </div>
    </footer>
  )
}
