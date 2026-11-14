"use client"

import { createColumnHelper } from "@tanstack/react-table"

import { type DataTableFeatures } from "./data-table-features"
import { LeaderboardEntry, ScrapeEntry } from "@/lib/models"
import { Button } from "../ui/button"
import { ArrowUpDownIcon } from "lucide-react"
import { Checkbox } from "../ui/checkbox"
import { formatDate, formatDecimal, formatLarge, formatPercent, formatSeconds } from "./format"

// Use `accessor` for data columns and `display` for columns without one.
const columnHelper = createColumnHelper<DataTableFeatures, LeaderboardEntry>()

const numberFmt = new Intl.NumberFormat("en-US")
const dateFmt = new Intl.DateTimeFormat("en-US", {
    timeZone: "America/New_York",
    timeZoneName: "shortGeneric"
})

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

export const columns = columnHelper.columns([
    columnHelper.accessor("rank", {
        header: ({ column }) => SortedColumn({ column, text: "Rank" })
    }),
    columnHelper.accessor("uid", {
        header: "UID",
    }),
    columnHelper.accessor("total_points", {
        header: ({ column }) => SortedColumn({ column, text: "Total pts" }),
        cell: ({ row }) => formatLarge(row.getValue("total_points"))
    }),
    columnHelper.accessor("solved", {
        header: ({ column }) => SortedColumn({ column, text: "Solved" }),
        cell: ({ row }) => formatLarge(row.getValue("solved"))
    }),
    columnHelper.accessor("attempted", {
        header: ({ column }) => SortedColumn({ column, text: "Attempted" }),
        cell: ({ row }) => formatLarge(row.getValue("attempted"))
    }),
    columnHelper.accessor("acceptance_rate", {
        header: ({ column }) => SortedColumn({ column, text: "Acceptance" }),
        cell: ({ row }) => formatPercent(row.getValue("acceptance_rate"))
    }),
    columnHelper.accessor("avg_time_s", {
        header: ({ column }) => SortedColumn({ column, text: "Avg time" }),
        cell: ({ row }) => formatSeconds(row.getValue("avg_time_s"))
    }),
    columnHelper.accessor("avg_points", {
        header: ({ column }) => SortedColumn({ column, text: "Avg pts" }),
        cell: ({ row }) => formatDecimal(row.getValue("avg_points"))
    }),
    // columnHelper.accessor("last_active", {
    //     header: ({ column }) => SortedColumn({ column, text: "Last active" }),
    //     cell: ({ row }) => formatDate(row.getValue("last_active"))
    // }),
])
