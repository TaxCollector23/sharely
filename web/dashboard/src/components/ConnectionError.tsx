import { RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'

export function ConnectionError({ onRetry }: { onRetry: () => void }) {
  return (
    <div className="flex flex-1 flex-col items-center justify-center px-6 py-24 text-center">
      <p className="text-base font-medium text-foreground">
        Can't reach Sharely.
      </p>
      <p className="mt-2 max-w-sm text-sm text-muted-foreground">
        The dashboard couldn't reach the local Sharely daemon. Make sure{' '}
        <code className="rounded border border-border bg-muted px-1.5 py-0.5 font-mono text-[13px] text-foreground">
          sharely
        </code>{' '}
        is running, then try again.
      </p>
      <Button variant="outline" size="sm" className="mt-5" onClick={onRetry}>
        <RefreshCw className="size-3.5" aria-hidden="true" />
        Retry
      </Button>
    </div>
  )
}
