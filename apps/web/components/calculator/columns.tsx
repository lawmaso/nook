/**
 * @file columns.tsx
 * 
 * Data table column definitions
 */

"use client"
import { createColumnHelper } from "@tanstack/react-table"
import { ArrowUpDownIcon } from "lucide-react"
import { LeaderboardEntry } from "@/lib/models"
import { Button } from "../ui/button"
import { type DataTableFeatures } from "./data-table-features"
import { formatDecimal, formatLarge, formatPercent, formatSeconds } from "./format"

// Use `accessor` for data columns and `display` for columns without one.
const columnHelper = createColumnHelper<DataTableFeatures, LeaderboardEntry>()

function SortedColumn({
    column,
    text
}: { column: any, text: string }) {
    return (
        <Button
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            variant="ghost"
            className="p-1"
        >
            {text}
            <ArrowUpDownIcon />
        </Button>
    )
}

export const columnLabels: Record<string, string> = {
    rank: "Rank",
    username: "Username",
    total_points: "Total pts",
    solved_attempted: "Solved",
    acceptance_rate: "Acceptance",
    avg_time_s: "Avg time",
    avg_points: "Avg pts",
}

export const columns = columnHelper.columns([
    columnHelper.accessor("rank", {
        header: ({ column }) => SortedColumn({ column, text: columnLabels[column.id] }),
        cell: ({ row }) => {
            const rank = row.original.rank
            const onPodium = rank <= 3

            return (
                <span className={onPodium
                    ? "inline-flex size-6 items-center justify-center bg-chart-3 font-semibold tabular-nums text-primary-foreground"
                    : "inline-flex size-6 items-center justify-center tabular-nums text-muted-foreground"
                }>
                    {rank}
                </span>
            )
        }
    }),
    columnHelper.accessor((e) => e.user.username, {
        id: "username",
        header: ({ column }) => SortedColumn({ column, text: columnLabels[column.id] }),
        cell: ({ getValue }) => <span className="font-medium">{getValue()}</span>
    }),
    columnHelper.accessor("total_points", {
        header: ({ column }) => SortedColumn({ column, text: columnLabels[column.id] }),
        cell: ({ row }) => <span className="tabular-nums">{formatLarge(row.getValue("total_points"))}</span>
    }),
    columnHelper.accessor((e) => `${e.solved}/${e.attempted}`, {
        id: "solved_attempted",
        header: ({ column }) => SortedColumn({ column, text: columnLabels[column.id] }),
        cell: ({ getValue }) => <span className="tabular-nums">{getValue()}</span>
    }),
    columnHelper.accessor("acceptance_rate", {
        header: ({ column }) => SortedColumn({ column, text: columnLabels[column.id] }),
        cell: ({ row }) => <span className="tabular-nums">{formatPercent(row.getValue("acceptance_rate"))}</span>
    }),
    columnHelper.accessor("avg_time_s", {
        header: ({ column }) => SortedColumn({ column, text: columnLabels[column.id] }),
        cell: ({ row }) => <span className="tabular-nums">{formatSeconds(row.getValue("avg_time_s"))}</span>
    }),
    columnHelper.accessor("avg_points", {
        header: ({ column }) => SortedColumn({ column, text: columnLabels[column.id] }),
        cell: ({ row }) => <span className="tabular-nums">{formatDecimal(row.getValue("avg_points"))}</span>
    }),
])
