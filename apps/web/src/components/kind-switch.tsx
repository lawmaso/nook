import Link from "next/link"

import { KIND_LABELS, KINDS, type Kind } from "@/lib/leaderboards"
import { cn } from "@/lib/utils"

/** Segmented control of links, so switching boards is a navigation. */
export function KindSwitch({ active, hrefFor }: { active: Kind; hrefFor: (kind: Kind) => string }) {
  return (
    <nav aria-label="Leaderboard" className="inline-flex h-8 items-center rounded-lg bg-muted p-[3px] text-muted-foreground">
      {KINDS.map((kind) => (
        <Link
          key={kind}
          href={hrefFor(kind)}
          aria-current={kind === active ? "page" : undefined}
          className={cn(
            "inline-flex h-full items-center rounded-md px-3 text-sm font-medium transition-colors hover:text-foreground",
            "aria-[current=page]:bg-background aria-[current=page]:text-foreground aria-[current=page]:shadow-sm dark:aria-[current=page]:bg-input/30",
          )}
        >
          {KIND_LABELS[kind]}
        </Link>
      ))}
    </nav>
  )
}
