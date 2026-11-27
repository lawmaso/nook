"use client"

import { format, parse, subMonths } from "date-fns"
import { CalendarIcon } from "lucide-react"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import { useMemo, useState, useTransition } from "react"
import type { DateRange } from "react-day-picker"

import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Separator } from "@/components/ui/separator"
import { formatDayLabel, rangePresets, toDay, type DayRange } from "@/lib/dates"
import { cn } from "@/lib/utils"

const DAY = "yyyy-MM-dd"
const REF = new Date(0)

function toDate(day: string) {
  return parse(day, DAY, REF)
}

/** Range picker that stores the selection in ?from=&to= (yyyy-MM-dd, site time). */
export function DateRangePicker({ range, className }: { range: DayRange; className?: string }) {
  const router = useRouter()
  const pathname = usePathname()
  const searchParams = useSearchParams()
  const [open, setOpen] = useState(false)
  const [draft, setDraft] = useState<DateRange | undefined>()
  const [pending, startTransition] = useTransition()
  const presets = useMemo(() => (open ? rangePresets() : []), [open])

  const selected: DateRange = draft ?? { from: toDate(range.from), to: toDate(range.to) }

  function apply(next: DayRange) {
    const params = new URLSearchParams(searchParams)
    params.set("from", next.from)
    params.set("to", next.to)
    setOpen(false)
    setDraft(undefined)
    startTransition(() => router.replace(`${pathname}?${params}`, { scroll: false }))
  }

  const label =
    range.from === range.to
      ? formatDayLabel(range.from)
      : `${format(toDate(range.from), "MMM d")} – ${formatDayLabel(range.to)}`

  return (
    <Popover
      open={open}
      onOpenChange={(next) => {
        setOpen(next)
        if (!next) setDraft(undefined)
      }}
    >
      <PopoverTrigger
        render={
          <Button variant="outline" className={cn("justify-start font-normal", pending && "opacity-70", className)} />
        }
      >
        <CalendarIcon />
        {label}
      </PopoverTrigger>
      <PopoverContent align="end" className="w-auto flex-col gap-0 p-0 sm:flex-row">
        <div className="flex flex-row flex-wrap gap-1 p-2 sm:w-36 sm:flex-col">
          {presets.map((p) => (
            <Button
              key={p.label}
              variant={p.range.from === range.from && p.range.to === range.to ? "secondary" : "ghost"}
              size="sm"
              className="justify-start"
              onClick={() => apply(p.range)}
            >
              {p.label}
            </Button>
          ))}
        </div>
        <Separator orientation="vertical" className="hidden sm:block" />
        <Calendar
          mode="range"
          numberOfMonths={2}
          // Two months are shown; end on the month of the range end (usually the current one).
          defaultMonth={subMonths(selected.to ?? selected.from ?? new Date(), 1)}
          selected={selected}
          disabled={{ after: new Date() }}
          onSelect={(_, day) => {
            // First click starts a fresh range; second click (any order) completes it.
            if (!draft?.from) {
              setDraft({ from: day, to: undefined })
              return
            }
            const [a, b] = draft.from <= day ? [draft.from, day] : [day, draft.from]
            apply({ from: toDay(a), to: toDay(b) })
          }}
        />
      </PopoverContent>
    </Popover>
  )
}
