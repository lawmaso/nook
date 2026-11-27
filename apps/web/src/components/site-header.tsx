import Link from "next/link"
import { Suspense } from "react"

import { ModeToggle } from "@/components/mode-toggle"
import { NavLinks, NavLinksFallback } from "@/components/nav-links"
import { UserSearch } from "@/components/user-search"

export function SiteHeader() {
  return (
    <header className="sticky top-0 z-40 border-b bg-background/80 backdrop-blur">
      <div className="mx-auto flex h-14 max-w-7xl items-center gap-2 px-4">
        <Link href="/" className="mr-2 font-heading text-lg font-semibold tracking-tight">
          Yagami
        </Link>
        <Suspense fallback={<NavLinksFallback />}>
          <NavLinks />
        </Suspense>
        <div className="ml-auto flex items-center gap-2">
          <UserSearch />
          <ModeToggle />
        </div>
      </div>
    </header>
  )
}
