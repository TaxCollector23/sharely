import { Copy, ExternalLink } from 'lucide-react'
import { toast } from 'sonner'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { api, type ShareDTO } from '@/lib/api'
import { formatTimestamp } from '@/lib/format'

interface QrDialogProps {
  share: ShareDTO | null
  onOpenChange: (open: boolean) => void
}

export function QrDialog({ share, onOpenChange }: QrDialogProps) {
  if (!share) return null

  const copyLink = async () => {
    await navigator.clipboard.writeText(share.primaryUrl)
    toast('Link copied')
  }

  return (
    <Dialog open={!!share} onOpenChange={onOpenChange}>
      <DialogContent className="w-[calc(100%-2rem)] sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{share.name}</DialogTitle>
        </DialogHeader>

        <div className="flex min-w-0 justify-center rounded-md border border-border bg-white p-4">
          <img
            src={api.qrSvgUrl(share.id)}
            alt={`QR code for ${share.primaryUrl}`}
            width={280}
            height={280}
            className="h-auto w-full max-w-[280px] min-w-0"
          />
        </div>

        <div className="flex items-center gap-2 rounded-md border border-border bg-muted/50 px-3 py-2">
          <code className="min-w-0 flex-1 truncate font-mono text-[13px] text-foreground">
            {share.primaryUrl}
          </code>
          <Button
            variant="ghost"
            size="icon"
            className="size-10 shrink-0"
            onClick={copyLink}
            aria-label="Copy link"
          >
            <Copy className="size-4" aria-hidden="true" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="size-10 shrink-0"
            asChild
          >
            <a
              href={share.primaryUrl}
              target="_blank"
              rel="noreferrer"
              aria-label="Open in new tab"
            >
              <ExternalLink className="size-4" aria-hidden="true" />
            </a>
          </Button>
        </div>

        <details className="group rounded-md border border-border">
          <summary className="cursor-pointer select-none list-none px-3 py-2 text-sm font-medium text-foreground [&::-webkit-details-marker]:hidden">
            Details
          </summary>
          <dl className="grid grid-cols-[auto,1fr] gap-x-3 gap-y-1.5 border-t border-border px-3 py-3 text-sm">
            <dt className="text-muted-foreground">Target</dt>
            <dd className="truncate font-mono text-[13px]">{share.target}</dd>
            <dt className="text-muted-foreground">Network</dt>
            <dd className="truncate font-mono text-[13px]">
              {share.networkAddress}
            </dd>
            <dt className="text-muted-foreground">Local hostname</dt>
            <dd className="truncate font-mono text-[13px]">
              {share.localHostname || '—'}
            </dd>
            <dt className="text-muted-foreground">Password</dt>
            <dd>{share.hasPassword ? 'Enabled' : 'None'}</dd>
            <dt className="text-muted-foreground">Created</dt>
            <dd>{formatTimestamp(share.createdAt)}</dd>
            <dt className="text-muted-foreground">Expires</dt>
            <dd>
              {share.expiresAt ? formatTimestamp(share.expiresAt) : 'Until stopped'}
            </dd>
          </dl>
        </details>
      </DialogContent>
    </Dialog>
  )
}
