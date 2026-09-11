import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'

interface StopDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  shareName: string
  onConfirm: () => void
  stopping: boolean
}

export function StopDialog({
  open,
  onOpenChange,
  shareName,
  onConfirm,
  stopping,
}: StopDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle>Stop sharing?</DialogTitle>
          <DialogDescription>
            Anyone using this link will lose access to{' '}
            <span className="font-mono text-foreground">{shareName}</span>.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="mt-2 gap-2 sm:gap-2">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={stopping}
          >
            Cancel
          </Button>
          <Button variant="destructive" onClick={onConfirm} disabled={stopping}>
            {stopping ? 'Stopping…' : 'Stop sharing'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
