import { columns } from "./columns"
import { DataTable } from "./data-table"
import { getLatestScrapeEntries } from "@/lib/queries"

export default async function Calculator() {
    const data = await getLatestScrapeEntries("monthly")
    console.log(data)
    console.log(data!.length)

    return (
        <div className="container mx-auto py-10">
            <DataTable columns={columns} data={data!} />
        </div>
    )
}