"use client"

import { createColumnHelper } from "@tanstack/react-table"

import { type DataTableFeatures } from "./data-table-features"
import { ScrapeEntry } from "@/lib/models"

// Use `accessor` for data columns and `display` for columns without one.
const columnHelper = createColumnHelper<DataTableFeatures, ScrapeEntry>()

export const columns = columnHelper.columns([
    columnHelper.accessor("rank", {
        header: "Rank",
    }),
    columnHelper.accessor("uid", {
        header: "UID",
    }),
    columnHelper.accessor("total_points", {
        header: "Total points",
    }),
    columnHelper.accessor("solved", {
        header: "Solved"
    }),
    columnHelper.accessor("attempted", {
        header: "Attempted"
    }),
    columnHelper.accessor("acceptance_rate", {
        header: "Acceptance"
    }),
    columnHelper.accessor("avg_time_s", {
        header: "Avg time"
    }),
    columnHelper.accessor("avg_points", {
        header: "Avg points"
    }),
    columnHelper.accessor("last_active", {
        header: "Last active"
    }),
])