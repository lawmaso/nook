// Port of apps/desktop/ui/utils/formatters.py

const large = new Intl.NumberFormat("en-US", { maximumFractionDigits: 2 })
const whole = new Intl.NumberFormat("en-US", { maximumFractionDigits: 0 })
const compact = new Intl.NumberFormat("en-US", {
  notation: "compact",
  maximumFractionDigits: 1,
})

export function formatLarge(n: number): string {
  return large.format(n)
}

export function formatCompact(n: number): string {
  return compact.format(n)
}

export function pointGap(gap: number): string {
  return formatLarge(gap)
}

/** "perfect-expected" puzzle estimate, or ∞ when it can't be closed. */
export function puzzleGap([perfect, expected]: readonly [number, number]): string {
  const fmt = (n: number) => (Number.isFinite(n) ? whole.format(n) : "∞")
  if (perfect >= 1e6 && expected >= 1e6) return "∞"
  return `${fmt(perfect)}-${fmt(expected)}`
}

export function timeGap(totalSeconds: number): string {
  if (!Number.isFinite(totalSeconds)) return "∞"
  const seconds = Math.max(0, Math.trunc(totalSeconds))

  const y = Math.floor(seconds / (3600 * 24 * 365))
  const d = Math.floor(seconds / (3600 * 24)) % 365
  const h = Math.floor(seconds / 3600) % 24
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60

  let out = ""
  if (y > 0) out += `${y}y `
  if (d > 0) out += `${d}d `
  if (h > 0) out += `${h}h `
  if (m > 0) out += `${m}m `
  return out + `${s}s`
}

export function secondsLabel(sec: number): string {
  return `${formatLarge(sec)}s`
}

export function percent(pct: number): string {
  return `${formatLarge(pct)}%`
}
