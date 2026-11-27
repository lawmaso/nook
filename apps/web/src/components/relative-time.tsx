"use client"

import { formatDistanceToNowStrict } from "date-fns"
import { useCallback, useSyncExternalStore } from "react"

import { timeGap } from "@/lib/format"

/** Current time quantized to `intervalMs`; null during SSR and hydration. */
function useNow(intervalMs: number): number | null {
  const subscribe = useCallback(
    (onTick: () => void) => {
      const id = setInterval(onTick, intervalMs)
      return () => clearInterval(id)
    },
    [intervalMs],
  )
  return useSyncExternalStore(
    subscribe,
    () => Math.floor(Date.now() / intervalMs) * intervalMs,
    () => null,
  )
}

/** "12 minutes ago", refreshed every 30s; renders nothing until mounted. */
export function RelativeTime({ iso, className }: { iso: string; className?: string }) {
  const now = useNow(30_000)
  return (
    <time dateTime={iso} title={new Date(iso).toLocaleString()} className={className} suppressHydrationWarning>
      {now === null ? " " : formatDistanceToNowStrict(new Date(iso), { addSuffix: true })}
    </time>
  )
}

/** Desktop-style "1d 3h 2m 5s" ticking every second. */
export function ElapsedTicker({ iso, className }: { iso: string; className?: string }) {
  const now = useNow(1000)
  if (now === null) return <span className={className}>…</span>
  return <span className={className}>{timeGap((now - Date.parse(iso)) / 1000)}</span>
}
