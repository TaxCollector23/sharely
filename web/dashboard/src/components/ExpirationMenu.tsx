import { Check, Clock } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import type { ShareDuration } from '@/lib/api'

const OPTIONS: { value: ShareDuration; label: string }[] = [
  { value: '15m', label: '15 minutes' },
  { value: '1h', label: '1 hour' },
  { value: '4h', label: '4 hours' },
  { value: 'forever', label: 'Until I stop it' },
]

interface ExpirationMenuProps {
  value: ShareDuration
  onChange: (duration: ShareDuration) => void
  disabled?: boolean
}

export function ExpirationMenu({ value, onChange, disabled }: ExpirationMenuProps) {
  const current = OPTIONS.find((o) => o.value === value)

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" disabled={disabled}>
          <Clock className="size-3.5" aria-hidden="true" />
          {current?.label ?? 'Change time'}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start">
        {OPTIONS.map((opt) => (
          <DropdownMenuItem
            key={opt.value}
            onSelect={() => onChange(opt.value)}
            className="justify-between"
          >
            {opt.label}
            {opt.value === value && (
              <Check className="size-3.5" aria-hidden="true" />
            )}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
