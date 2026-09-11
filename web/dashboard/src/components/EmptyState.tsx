export function EmptyState() {
  return (
    <div className="flex flex-1 flex-col items-center justify-center px-6 py-20">
      <div className="w-full max-w-sm text-center">
        <p className="text-[15px] font-medium text-foreground">
          Nothing is being shared yet
        </p>
        <p className="mt-1.5 text-sm text-muted-foreground">
          Run this from any folder to share it on your network.
        </p>
        <div className="mt-5 flex items-center gap-2.5 rounded-md border border-border bg-muted/40 px-4 py-3 text-left">
          <span
            className="select-none font-mono text-sm text-muted-foreground"
            aria-hidden="true"
          >
            $
          </span>
          <code className="font-mono text-sm text-foreground">
            sharely start
          </code>
        </div>
      </div>
    </div>
  )
}
