"use client"

import { useState } from "react"
import { cn } from "cn"
import {
    useTable,
    type ColumnDef,
    type RowData,
    type SortingState,
    type ColumnFiltersState,
    type ColumnVisibilityState,
    RowSelectionState
} from "@tanstack/react-table"

import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table"
import { Input } from "@/components/ui/input"
import { DropdownMenu, DropdownMenuCheckboxItem, DropdownMenuContent, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"

import { features, type DataTableFeatures } from "./data-table-features"
import { columnLabels } from "./columns"

interface DataTableProps<TData extends RowData> {
    columns: ColumnDef<DataTableFeatures, TData>[]
    data: TData[]
}

function rowTone(rank: number | undefined) {
    if (rank === 1) return "bg-primary/20 hover:bg-primary/25 border-1-4 border-1-primary"
    if (rank === 2) return "bg-char-2/20 hover:bg-chart-2/25 border-1-4 border-1-chart-2"
    return "odd:bg-card even:bg-muted/50 hover:bg-accent/60 border-1-4 border-1-transparent"
}

export function DataTable<TData extends RowData>({
    columns,
    data,
}: DataTableProps<TData>) {
    const [sorting, setSorting] = useState<SortingState>([])
    const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([])
    const [columnVisibility, setColumnVisibility] = useState<ColumnVisibilityState>({})
    const [rowSelection, setRowSelection] = useState<RowSelectionState>({})

    const table = useTable({
        features,
        data,
        columns,
        manualPagination: true,
        onSortingChange: setSorting,
        onColumnFiltersChange: setColumnFilters,
        onColumnVisibilityChange: setColumnVisibility,
        onRowSelectionChange: setRowSelection,
        state: {
            sorting,
            columnFilters,
            columnVisibility,
            rowSelection
        },
    })

    return (
        <div className="mx-auto flex h-full min-h-0 w-full max-w-5xl flex-col flex-1 overflow-hidden border bg-card shadow-sm">
            <div className="flex items-center gap-2 border-b bg-secondary/40 p-3">
                <Input
                    placeholder="Search username"
                    value={(table.getColumn("username")?.getFilterValue() as string) ?? ""}
                    onChange={(event) =>
                        table.getColumn("username")?.setFilterValue(event.target.value)
                    }
                    className="max-w-xs bg-card"
                    autoComplete="off"
                    autoCapitalize="off"
                />
                <DropdownMenu>
                    <DropdownMenuTrigger render={<Button variant="outline" className="ml-auto bg-card" />}>
                        Columns
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                        {table
                            .getAllColumns()
                            .filter((column) => column.getCanHide())
                            .map((column) => (
                                <DropdownMenuCheckboxItem
                                    key={column.id}
                                    checked={column.getIsVisible()}
                                    onCheckedChange={(value) => column.toggleVisibility(!!value)}
                                >
                                    {columnLabels[column.id] ?? column.id}
                                </DropdownMenuCheckboxItem>
                            ))}
                    </DropdownMenuContent>
                </DropdownMenu>
            </div>
            <ScrollArea className="min-h-0 flex-1">
                <Table>
                    <TableHeader className="sticky top-0 z-10 bg-card">
                        {table.getHeaderGroups().map((headerGroup) => (
                            <TableRow key={headerGroup.id} className="hover:bg-transparent">
                                {headerGroup.headers.map((header) => (
                                    <TableHead key={header.id} className="h-11 first:pl-5">
                                        {header.isPlaceholder ? null : (
                                            <table.FlexRender header={header} />
                                        )}
                                    </TableHead>
                                ))}
                            </TableRow>
                        ))}
                    </TableHeader>
                    <TableBody>
                        {table.getRowModel().rows?.length ? (
                            table.getRowModel().rows.map((row) => (
                                <TableRow
                                    key={row.id}
                                    data-state={row.getIsSelected() && "selected"}
                                    className={cn(
                                        "transition-colors",
                                        // rowTone((row.original as { rank?: number }).rank)
                                    )}
                                >
                                    {row.getVisibleCells().map((cell) => (
                                        <TableCell key={cell.id} className="py-3 first:pl-4">
                                            <table.FlexRender cell={cell} />
                                        </TableCell>
                                    ))}
                                </TableRow>
                            ))
                        ) : (
                            <TableRow>
                                <TableCell colSpan={columns.length} className="h-24 text-center text-muted-foreground">
                                    No results.
                                </TableCell>
                            </TableRow>
                        )}
                    </TableBody>
                </Table>
            </ScrollArea>
        </div>
    )
}