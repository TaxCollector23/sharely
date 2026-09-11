import { useCallback, useEffect, useRef, useState } from 'react'

export type ShareType = 'file' | 'directory' | 'website' | 'proxy'
export type ShareStatus = 'running' | 'expired' | 'stopped'
export type ShareDuration = '15m' | '1h' | '4h' | 'forever'

export interface ShareDTO {
  id: string
  name: string
  target: string
  type: ShareType
  status: ShareStatus
  url: string
  primaryUrl: string
  remaining: string
  duration: ShareDuration
  createdAt: string
  expiresAt: string | null
  hasPassword: boolean
  deviceCount: number
  lastAccess: string | null
  networkAddress: string
  localHostname: string
}

export interface StatusDTO {
  version: string
  startedAt: string
  network: string
  localHost: string
  interface: string
  activeCount: number
}

export interface LogEntry {
  time: string
  method: string
  path: string
  status: number
}

class ApiError extends Error {
  status?: number

  constructor(message: string, status?: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let res: Response
  try {
    res = await fetch(path, {
      ...init,
      headers: {
        ...(init?.body ? { 'Content-Type': 'application/json' } : {}),
        ...init?.headers,
      },
    })
  } catch {
    throw new ApiError("Can't reach Sharely")
  }
  if (!res.ok) {
    const text = await res.text().catch(() => '')
    throw new ApiError(text || `Request failed (${res.status})`, res.status)
  }
  if (res.status === 204) return undefined as T
  const contentType = res.headers.get('content-type') ?? ''
  if (contentType.includes('application/json')) {
    return (await res.json()) as T
  }
  return (await res.text()) as unknown as T
}

export const api = {
  getStatus: () => request<StatusDTO>('/api/status'),
  getShares: () => request<ShareDTO[]>('/api/shares'),
  getShare: (id: string) => request<ShareDTO>(`/api/shares/${id}`),
  stopShare: (id: string) =>
    request<ShareDTO>(`/api/shares/${id}/stop`, { method: 'POST' }),
  setExpiration: (id: string, duration: ShareDuration) =>
    request<ShareDTO>(`/api/shares/${id}/expiration`, {
      method: 'POST',
      body: JSON.stringify({ duration }),
    }),
  getLogs: (id: string) => request<LogEntry[]>(`/api/shares/${id}/logs`),
  qrSvgUrl: (id: string) => `/api/shares/${id}/qr.svg`,
}

export { ApiError }

/**
 * Polls GET /api/shares on an interval, keeping remaining-time and device
 * counts fresh. Also tracks whether the daemon is reachable at all, so the
 * UI can show a calm "can't reach Sharely" state instead of a blank screen
 * (mainly relevant to `npm run dev` against a stopped daemon).
 */
export function useShares(intervalMs = 3000) {
  const [shares, setShares] = useState<ShareDTO[] | null>(null)
  const [error, setError] = useState<Error | null>(null)
  const [loading, setLoading] = useState(true)
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const inFlight = useRef(false)

  const refresh = useCallback(async () => {
    if (inFlight.current) return
    inFlight.current = true
    try {
      const data = await api.getShares()
      setShares(data)
      setError(null)
    } catch (err) {
      setError(err as Error)
    } finally {
      inFlight.current = false
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    refresh()
    timer.current = setInterval(refresh, intervalMs)
    return () => {
      if (timer.current) clearInterval(timer.current)
    }
  }, [refresh, intervalMs])

  return { shares, error, loading, refresh }
}
