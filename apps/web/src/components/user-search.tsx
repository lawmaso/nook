"use client"

import { SearchIcon, UserIcon } from "lucide-react"
import { useRouter } from "next/navigation"
import { useEffect, useState, useTransition } from "react"

import { Button } from "@/components/ui/button"
import {
  Command,
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
import { searchUsersAction } from "@/lib/actions"
import type { UserHit } from "@/lib/queries"

export function UserSearch() {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState("")
  const [hits, setHits] = useState<UserHit[]>([])
  const [pending, startTransition] = useTransition()

  useEffect(() => {
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        setOpen((o) => !o)
      }
    }
    window.addEventListener("keydown", onKeyDown)
    return () => window.removeEventListener("keydown", onKeyDown)
  }, [])

  useEffect(() => {
    const q = query.trim()
    if (!q) return
    const id = setTimeout(() => {
      startTransition(async () => setHits(await searchUsersAction(q)))
    }, 200)
    return () => clearTimeout(id)
  }, [query])

  function go(uid: number) {
    setOpen(false)
    setQuery("")
    router.push(`/u/${uid}`)
  }

  const visible = query.trim() ? hits : []

  return (
    <>
      <Button
        variant="outline"
        className="w-9 justify-start px-0 text-muted-foreground sm:w-56 sm:px-2.5"
        onClick={() => setOpen(true)}
        aria-label="Search users"
      >
        <SearchIcon className="mx-auto sm:mx-0" />
        <span className="hidden sm:inline">Search users…</span>
        <kbd className="ml-auto hidden rounded border bg-muted px-1.5 font-mono text-[10px] sm:inline">⌘K</kbd>
      </Button>
      <CommandDialog open={open} onOpenChange={setOpen} title="Search users" description="Find a user by username">
        <Command shouldFilter={false}>
          <CommandInput placeholder="Username…" value={query} onValueChange={setQuery} />
          <CommandList>
            {query.trim() && !pending && <CommandEmpty>No users found.</CommandEmpty>}
            {visible.length > 0 && (
              <CommandGroup heading="Users">
                {visible.map((u) => (
                  <CommandItem key={u.uid} value={String(u.uid)} onSelect={() => go(u.uid)}>
                    <UserIcon />
                    {u.username}
                  </CommandItem>
                ))}
              </CommandGroup>
            )}
          </CommandList>
        </Command>
      </CommandDialog>
    </>
  )
}
