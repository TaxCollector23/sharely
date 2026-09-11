const facts = [
  'No upload',
  'No cloud',
  'No account',
  'No API key',
]

export function Security() {
  return (
    <section className="border-t border-[var(--border)] bg-[var(--bg-sunken)] py-16">
      <div className="mx-auto max-w-5xl px-5 sm:px-6">
        <div className="grid gap-8 sm:grid-cols-[1.2fr_1fr] sm:items-start">
          <div>
            <h2 className="text-xl font-semibold tracking-tight sm:text-2xl">
              Your files stay on your local network.
            </h2>
            <p className="mt-3 max-w-md text-[15px] leading-relaxed text-[var(--text-muted)]">
              Sharely runs a server on your machine and answers requests from other devices on the
              same Wi-Fi or LAN. There&rsquo;s no relay in between, and this version of Sharely
              doesn&rsquo;t open a tunnel to the public internet.
            </p>
            <p className="mt-3 max-w-md text-[15px] leading-relaxed text-[var(--text-muted)]">
              Every share expires on its own &mdash; 15 minutes, 1 hour, 4 hours, or until you stop
              it &mdash; and you can add a password if the network isn&rsquo;t one you fully trust.
            </p>
            <p className="mt-3 max-w-md text-sm leading-relaxed text-[var(--text-faint)]">
              Sharely also checks that the link it hands you will actually work: it confirms{' '}
              <code className="font-mono">sharely.local</code> resolves before using it, falls back
              to your LAN IP if it doesn&rsquo;t, and offers a{' '}
              <code className="font-mono">127.0.0.1</code> link if the LAN address isn&rsquo;t
              reachable either.
            </p>
          </div>
          <ul className="grid grid-cols-2 gap-px overflow-hidden rounded-lg border border-[var(--border-strong)] bg-[var(--border-strong)] sm:gap-0">
            {facts.map((fact) => (
              <li
                key={fact}
                className="bg-[var(--bg-raised)] px-5 py-6 text-[15px] font-medium"
              >
                {fact}
              </li>
            ))}
          </ul>
        </div>
      </div>
    </section>
  )
}
