"use client"

import {
  createColumnHelper,
  createSortedRowModel,
  rowSortingFeature,
  sortFn_alphanumeric,
  sortFn_basic,
  tableFeatures,
  useTable,
} from "@tanstack/react-table"
import { ArrowDownIcon, ArrowUpDownIcon, ArrowUpIcon, SearchIcon } from "lucide-react"
import Link from "next/link"
import { useDeferredValue, useMemo, useState } from "react"

import { InputGroup, InputGroupAddon, InputGroupInput } from "@/components/ui/input-group"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { formatLarge, percent, secondsLabel } from "@/lib/format"
import type { LeaderboardEntry } from "@/lib/leaderboards"
import { cn } from "@/lib/utils"

const features = tableFeatures({
  rowSortingFeature,
  sortedRowModel: createSortedRowModel(),
  sortFns: { alphanumeric: sortFn_alphanumeric, basic: sortFn_basic },
})

const helper = createColumnHelper<typeof features, LeaderboardEntry>()

const MEDALS: Record<number, string> = {
  1: "bg-amber-400/20 text-amber-700 dark:text-amber-300",
  2: "bg-zinc-400/20 text-zinc-700 dark:text-zinc-300",
  3: "bg-orange-400/20 text-orange-700 dark:text-orange-300",
}

// Same columns as apps/desktop/ui/theme/columns.py
const columns = helper.columns([
  helper.accessor("rank", {
    header: "#",
    sortFn: "basic",
    cell: ({ getValue }) => {
      const rank = getValue()
      return (
        <span
          className={cn(
            "inline-flex h-6 min-w-6 items-center justify-center rounded-md px-1.5 font-mono text-xs tabular-nums",
            MEDALS[rank] ?? "text-muted-foreground",
          )}
        >
          {rank}
        </span>
      )
    },
  }),
  helper.accessor("username", {
    header: "Username",
    sortFn: "alphanumeric",
    cell: ({ row, getValue }) => (
      <Link href={`/u/${row.original.uid}`} className="font-medium underline-offset-4 hover:underline">
        {getValue()}
      </Link>
    ),
  }),
  helper.accessor("total_points", {
    header: "Total pts",
    sortFn: "basic",
    sortDescFirst: true,
    cell: ({ getValue }) => formatLarge(getValue()),
  }),
  helper.accessor("solved", {
    header: "Solved",
    sortFn: "basic",
    sortDescFirst: true,
    cell: ({ row, getValue }) => (
      <>
        {formatLarge(getValue())}
        <span className="text-muted-foreground"> / {formatLarge(row.original.attempted)}</span>
      </>
    ),
  }),
  helper.accessor("avg_time_s", {
    header: "Avg time",
    sortFn: "basic",
    cell: ({ getValue }) => secondsLabel(getValue()),
  }),
  helper.accessor("acceptance_rate", {
    header: "Acceptance",
    sortFn: "basic",
    sortDescFirst: true,
    cell: ({ getValue }) => percent(getValue()),
  }),
])

const NUMERIC = new Set(["rank", "total_points", "solved", "avg_time_s", "acceptance_rate"])

export function LeaderboardTable({ entries }: { entries: LeaderboardEntry[] }) {
  const [filter, setFilter] = useState("")
  const deferredFilter = useDeferredValue(filter)

  const data = useMemo(() => {
    const q = deferredFilter.trim().toLowerCase()
    return q ? entries.filter((e) => e.username.toLowerCase().includes(q)) : entries
  }, [entries, deferredFilter])

  const table = useTable({ features, columns, data, initialState: { sorting: [{ id: "rank", desc: false }] } })
  const rows = table.getRowModel().rows

  return (
    <div className="flex flex-col gap-3">
      <InputGroup className="max-w-xs">
        <InputGroupInput
          placeholder="Filter usernames…"
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          aria-label="Filter usernames"
        />
        <InputGroupAddon>
          <SearchIcon />
        </InputGroupAddon>
      </InputGroup>

      <div className="overflow-hidden rounded-lg border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((group) => (
              <TableRow key={group.id}>
                {group.headers.map((header) => {
                  const sorted = header.column.getIsSorted()
                  const Icon = sorted === "asc" ? ArrowUpIcon : sorted === "desc" ? ArrowDownIcon : ArrowUpDownIcon
                  return (
                    <TableHead
                      key={header.id}
                      aria-sort={sorted === "asc" ? "ascending" : sorted === "desc" ? "descending" : "none"}
                      className={cn(header.column.id === "rank" && "w-14")}
                    >
                      <button
                        type="button"
                        onClick={header.column.getToggleSortingHandler()}
                        className={cn(
                          "-mx-1 inline-flex items-center gap-1 rounded px-1 py-0.5 hover:text-foreground",
                          sorted ? "text-foreground" : "text-muted-foreground",
                        )}
                      >
                        <table.FlexRender header={header} />
                        <Icon className={cn("size-3.5", !sorted && "opacity-40")} />
                      </button>
                    </TableHead>
                  )
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {rows.length === 0 ? (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">
                  No users match &ldquo;{filter}&rdquo;.
                </TableCell>
              </TableRow>
            ) : (
              rows.map((row) => (
                <TableRow key={row.id}>
                  {row.getAllCells().map((cell) => (
                    <TableCell key={cell.id} className={cn(NUMERIC.has(cell.column.id) && "tabular-nums")}>
                      <table.FlexRender cell={cell} />
                    </TableCell>
                  ))}
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
