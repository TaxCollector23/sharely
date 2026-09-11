export function Header() {
  return (
    <header className="border-b border-[var(--border)]">
      <div className="mx-auto flex max-w-5xl items-center justify-between px-5 py-4 sm:px-6">
        <a href="#top" className="flex items-center gap-2 text-[15px] font-semibold tracking-tight">
          <svg width="20" height="20" viewBox="0 0 32 32" fill="none" aria-hidden="true">
            <rect width="32" height="32" rx="7" fill="var(--accent)" />
            <circle cx="10" cy="16" r="2.4" fill="var(--accent-text-on)" />
            <circle cx="22" cy="9" r="2.4" fill="var(--accent-text-on)" />
            <circle cx="22" cy="23" r="2.4" fill="var(--accent-text-on)" />
            <path
              d="M12.5 14.5 19.5 10.5M12.5 17.5 19.5 21.5"
              stroke="var(--accent-text-on)"
              strokeWidth="1.6"
              strokeLinecap="round"
            />
          </svg>
          Sharely
        </a>
        <nav className="hidden items-center gap-6 text-sm text-[var(--text-muted)] sm:flex">
          <a href="#how-it-works" className="transition-colors hover:text-[var(--text)]">
            How it works
          </a>
          <a href="#cli" className="transition-colors hover:text-[var(--text)]">
            CLI
          </a>
          <a href="#install" className="transition-colors hover:text-[var(--text)]">
            Install
          </a>
        </nav>
        <div className="flex items-center gap-3">
          <a
            href="https://github.com/TaxCollector23/sharely" target="_blank" rel="noreferrer"
            className="hidden text-sm text-[var(--text-muted)] transition-colors hover:text-[var(--text)] sm:inline"
          >
            GitHub
          </a>
          <a
            href="#install"
            className="rounded-md bg-[var(--accent)] px-3.5 py-1.5 text-sm font-medium text-[var(--accent-text-on)] transition-colors hover:bg-[var(--accent-hover)]"
          >
            Get Sharely
          </a>
        </div>
      </div>
    </header>
  )
}
