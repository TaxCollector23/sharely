export function formatTimestamp(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}

const TYPE_LABELS: Record<string, string> = {
  file: 'File',
  directory: 'Folder',
  website: 'Website',
  proxy: 'Dev server',
}

export function shareTypeLabel(type: string): string {
  return TYPE_LABELS[type] ?? type
}

/** True once remaining/expiry implies under ~10 minutes are left. */
export function isRunningLow(expiresAt: string | null): boolean {
  if (!expiresAt) return false
  const ms = new Date(expiresAt).getTime() - Date.now()
  return ms > 0 && ms <= 10 * 60 * 1000
}
