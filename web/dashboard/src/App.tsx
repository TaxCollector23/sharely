import { useState } from 'react'
import { Share2 } from 'lucide-react'
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
              <div className="flex size-7 items-center justify-center rounded-md bg-primary text-primary-foreground">
                <Share2 className="size-4" aria-hidden="true" />
              </div>
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

        <main className="mx-auto flex w-full max-w-3xl flex-1 flex-col px-4 sm:px-6">
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
