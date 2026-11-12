const ACCEPTANCE_IS_FRACTION = false

const intFmt = new Intl.NumberFormat("en-US", { maximumFractionDigits: 0 })
const decimalFmt = new Intl.NumberFormat("en-US", { maximumFractionDigits: 1 })

export const formatLarge = (n: number) => intFmt.format(n)
export const formatDecimal = (n: number) => decimalFmt.format(n)

export function formatPercent(n: number) {
    const pct = ACCEPTANCE_IS_FRACTION ? n * 100 : n
    return `${pct.toFixed(2)}%`
}

/** Short seconds value, e.g. 42.5s or 3m 12s */
export function formatSeconds(totalSeconds: number) {
    if (totalSeconds < 60) return `${totalSeconds.toFixed(1)}s`
    const m = Math.floor(totalSeconds / 60)
    const s = Math.round(totalSeconds % 60)
    return `${m}m ${s}s`
}

/** Long duration, e.g. 2d 3h 4m 5s (leading zero units are dropped) */
export function formatDuration(totalSeconds: number) {
    const total = Math.max(0, Math.round(totalSeconds))
    const d = Math.floor(total / 86_400)
    const h = Math.floor((total % 86_400) / 3_600)
    const m = Math.floor((total % 3_600) / 60)
    const s = total % 60

    const parts: string[] = []
    if (d) parts.push(`${d}d`)
    if (d || h) parts.push(`${h}h`)
    if (d || h || m) parts.push(`${m}m`)
    parts.push(`${s}s`)
    return parts.join(" ")
}

const dateFmt = new Intl.DateTimeFormat("en-US", {
    timeZone: "America/New_York",
    timeZoneName: "shortGeneric",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
})

export function formatDate(raw: string) {
    const date = new Date(raw)
    if (Number.isNaN(date.getTime())) return "—"
    return dateFmt.format(date).replace(",", " @")
}
