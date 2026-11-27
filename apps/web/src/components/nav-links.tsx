"use client"

import { ChartLineIcon, TrophyIcon } from "lucide-react"
import Link from "next/link"
import { usePathname } from "next/navigation"

import { buttonVariants } from "@/components/ui/button"
import { cn } from "@/lib/utils"

const NAV = [
  { href: "/leaderboard/monthly", match: "/leaderboard", label: "Leaderboard", icon: TrophyIcon },
  { href: "/analytics", match: "/analytics", label: "Analytics", icon: ChartLineIcon },
] as const

function Links({ pathname }: { pathname: string | null }) {
  return (
    <nav className="flex items-center gap-1">
      {NAV.map(({ href, match, label, icon: Icon }) => (
        <Link
          key={href}
          href={href}
          aria-current={pathname?.startsWith(match) ? "page" : undefined}
          className={cn(
            buttonVariants({ variant: "ghost" }),
            "text-muted-foreground aria-[current=page]:bg-muted aria-[current=page]:text-foreground",
          )}
        >
          <Icon />
          <span className="hidden sm:inline">{label}</span>
        </Link>
      ))}
    </nav>
  )
}

export function NavLinks() {
  return <Links pathname={usePathname()} />
}

/** Prerendered fallback: same links, no active state. */
export function NavLinksFallback() {
  return <Links pathname={null} />
}
