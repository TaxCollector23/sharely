import { useState } from 'react'
import { Toaster } from '@/components/ui/sonner'
import { EmptyState } from '@/components/EmptyState'
import { ConnectionError } from '@/components/ConnectionError'
import { ShareCard } from '@/components/ShareCard'
import { QrDialog } from '@/components/QrDialog'
import { useShares, type ShareDTO } from '@/lib/api'
import { cn } from '@/lib/utils'

function App() {
  const { shares, error, loading, refresh } = useShares()
  const [qrShare, setQrShare] = useState<ShareDTO | null>(null)

  const running = shares?.filter((s) => s.status === 'running') ?? []
  const connected = shares !== null && !error
  // Keep the QR dialog's data fresh as polling updates the underlying share.
  const activeQrShare = qrShare
    ? (running.find((s) => s.id === qrShare.id) ?? qrShare)
    : null

  return (
    <>
      <div className="flex min-h-screen flex-col">
        <header className="border-b border-border">
          <div className="mx-auto flex max-w-3xl items-center justify-between px-4 py-4 sm:px-6">
            <div className="flex items-center gap-2">
              <svg className="dashboard-logo" width="40" height="40" viewBox="0 0 64 64" fill="none" aria-hidden="true">
                <rect x="1" y="1" width="62" height="62" rx="15" fill="var(--cream)" stroke="var(--ink)" strokeWidth="2" />
                <rect x="16.5" y="34" width="3" height="3.5" fill="var(--ink)" /><rect x="11" y="38.5" width="14" height="2.5" rx="1.25" fill="var(--ink)" />
                <rect x="4" y="12" width="28" height="22" rx="3" stroke="var(--ink)" strokeWidth="3" /><rect x="11.5" y="17" width="6" height="4" rx="1.5" fill="var(--ink)" /><rect x="11.5" y="20" width="13" height="9" rx="1.6" fill="var(--ink)" />
                <rect x="36" y="8" width="24" height="48" rx="4" stroke="var(--ink)" strokeWidth="3" /><rect x="47" y="52" width="3" height="1.8" rx=".9" fill="var(--ink)" /><rect x="41.5" y="23" width="6" height="4" rx="1.5" fill="var(--ink)" /><rect x="41.5" y="26" width="13" height="9.5" rx="1.6" fill="var(--ink)" />
              </svg>
              <span className="text-[15px] font-semibold tracking-tight text-foreground">
                Sharely
              </span>
            </div>

            <div
              className="flex items-center gap-1.5 text-xs font-medium text-muted-foreground"
              role="status"
              aria-live="polite"
            >
              <span
                className={cn(
                  'size-1.5 rounded-full',
                  connected
                    ? running.length > 0
                      ? 'animate-pulse bg-primary'
                      : 'bg-primary/60'
                    : 'bg-muted-foreground/40',
                )}
                aria-hidden="true"
              />
              {connected
                ? `${running.length} active`
                : 'Disconnected'}
            </div>
          </div>
        </header>

        <main className="dashboard-main mx-auto flex w-full max-w-3xl flex-1 flex-col px-4 sm:px-6">
          <section className="dashboard-hero">
            <span className="dashboard-kicker">LOCAL SHARING / CONTROL CENTER</span>
            <h1>Sharing<br /><em>Made Simple.</em></h1>
            <p>Manage the files and sites you’re sharing on your network.</p>
          </section>
          {error && !shares ? (
            <ConnectionError onRetry={refresh} />
          ) : loading && !shares ? null : running.length === 0 ? (
            <EmptyState />
          ) : (
            <section className="py-6">
              <h1 className="mb-3 text-sm font-medium text-muted-foreground">
                Active shares
              </h1>
              <ul className="flex flex-col gap-3">
                {running.map((share) => (
                  <ShareCard
                    key={share.id}
                    share={share}
                    onShowQr={setQrShare}
                    onChanged={refresh}
                  />
                ))}
              </ul>
            </section>
          )}
        </main>
      </div>

      <QrDialog
        share={activeQrShare}
        onOpenChange={(open) => !open && setQrShare(null)}
      />
      <Toaster position="bottom-center" />
    </>
  )
}

export default App
