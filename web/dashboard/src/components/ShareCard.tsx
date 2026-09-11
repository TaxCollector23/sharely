import { useState } from 'react'
import {
  Copy,
  ExternalLink,
  File,
  Folder,
  Globe,
  Lock,
  QrCode,
  Server,
  Square,
  Users,
} from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { ExpirationMenu } from '@/components/ExpirationMenu'
import { StopDialog } from '@/components/StopDialog'
import { api, type ShareDTO, type ShareDuration } from '@/lib/api'
import { isRunningLow, shareTypeLabel } from '@/lib/format'
import { cn } from '@/lib/utils'

const TYPE_ICONS: Record<string, typeof File> = {
  file: File,
  directory: Folder,
  website: Globe,
  proxy: Server,
}

interface ShareCardProps {
  share: ShareDTO
  onShowQr: (share: ShareDTO) => void
  onChanged: () => void
}

export function ShareCard({ share, onShowQr, onChanged }: ShareCardProps) {
  const [stopOpen, setStopOpen] = useState(false)
  const [stopping, setStopping] = useState(false)
  const [changingExpiration, setChangingExpiration] = useState(false)

  const TypeIcon = TYPE_ICONS[share.type] ?? File
  const low = isRunningLow(share.expiresAt)

  const copyLink = async () => {
    await navigator.clipboard.writeText(share.primaryUrl)
    toast('Link copied')
  }

  const handleStop = async () => {
    setStopping(true)
    try {
      await api.stopShare(share.id)
      setStopOpen(false)
      onChanged()
    } catch {
      toast.error("Couldn't stop the share")
    } finally {
      setStopping(false)
    }
  }

  const handleExpirationChange = async (duration: ShareDuration) => {
    setChangingExpiration(true)
    try {
      await api.setExpiration(share.id, duration)
      onChanged()
    } catch {
      toast.error("Couldn't change the expiration")
    } finally {
      setChangingExpiration(false)
    }
  }

  return (
    <li className="rounded-lg border border-border bg-card p-4 sm:p-5">
      <div className="flex items-start justify-between gap-4">
        <div className="flex min-w-0 items-center gap-3">
          <div className="flex size-9 shrink-0 items-center justify-center rounded-md bg-muted">
            <TypeIcon
              className="size-[18px] text-muted-foreground"
              aria-hidden="true"
            />
          </div>
          <div className="min-w-0">
            <div className="flex items-center gap-1.5">
              <span className="truncate text-[15px] font-semibold text-foreground">
                {share.name}
              </span>
              {share.hasPassword && (
                <Lock
                  className="size-3.5 shrink-0 text-muted-foreground"
                  aria-label="Password protected"
                />
              )}
            </div>
            <div className="mt-0.5 flex items-center gap-1.5 text-xs text-muted-foreground">
              <span>{shareTypeLabel(share.type)}</span>
              {share.deviceCount > 0 && (
                <>
                  <span aria-hidden="true">·</span>
                  <span className="inline-flex items-center gap-1">
                    <Users className="size-3" aria-hidden="true" />
                    {share.deviceCount}{' '}
                    {share.deviceCount === 1 ? 'device' : 'devices'}
                  </span>
                </>
              )}
            </div>
          </div>
        </div>

        <span
          className={cn(
            'shrink-0 text-xs font-medium text-muted-foreground',
            low && 'text-amber-600 dark:text-amber-500',
          )}
        >
          {share.remaining}
        </span>
      </div>

      <div className="mt-3 flex items-center gap-2 rounded-md border border-border bg-muted/40 py-1.5 pl-3 pr-1.5">
        <code className="min-w-0 flex-1 truncate font-mono text-[13px] text-foreground">
          {share.primaryUrl}
        </code>
        <Button
          variant="secondary"
          size="sm"
          className="h-10 shrink-0 gap-1.5 px-3 text-xs"
          onClick={copyLink}
        >
          <Copy className="size-3.5" aria-hidden="true" />
          Copy
        </Button>
      </div>

      <div className="mt-3 flex flex-wrap items-center gap-2">
        <Button size="sm" className="gap-1.5" asChild>
          <a href={share.primaryUrl} target="_blank" rel="noreferrer">
            <ExternalLink className="size-3.5" aria-hidden="true" />
            Open
          </a>
        </Button>

        <Button
          variant="outline"
          size="sm"
          className="gap-1.5"
          onClick={() => onShowQr(share)}
        >
          <QrCode className="size-3.5" aria-hidden="true" />
          QR code
        </Button>

        <ExpirationMenu
          value={share.duration}
          onChange={handleExpirationChange}
          disabled={changingExpiration}
        />

        <Button
          variant="outline"
          size="sm"
          className="ml-auto gap-1.5 text-muted-foreground hover:border-destructive/40 hover:text-destructive"
          onClick={() => setStopOpen(true)}
        >
          <Square className="size-3.5" aria-hidden="true" />
          Stop
        </Button>
      </div>

      <StopDialog
        open={stopOpen}
        onOpenChange={setStopOpen}
        shareName={share.name}
        onConfirm={handleStop}
        stopping={stopping}
      />
    </li>
  )
}
