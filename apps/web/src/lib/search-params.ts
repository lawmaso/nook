import type { Kind } from "@/lib/leaderboards"

type Params = Record<string, string | string[] | undefined>

export function first(value: string | string[] | undefined): string | undefined {
  return Array.isArray(value) ? value[0] : value
}

/** Build a link to `path` for `kind`, keeping an explicitly chosen date range. */
export function kindHref(path: string, kind: Kind, params: Params): string {
  const next = new URLSearchParams({ kind })
  const from = first(params.from)
  const to = first(params.to)
  if (from && to) {
    next.set("from", from)
    next.set("to", to)
  }
  return `${path}?${next}`
}
