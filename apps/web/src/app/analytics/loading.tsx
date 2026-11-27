import { Skeleton } from "@/components/ui/skeleton"

export default function Loading() {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-end justify-between">
        <div className="flex flex-col gap-2">
          <Skeleton className="h-8 w-40" />
          <Skeleton className="h-4 w-80" />
        </div>
        <Skeleton className="h-8 w-80" />
      </div>
      <div className="grid gap-6 lg:grid-cols-2">
        <Skeleton className="h-[420px] rounded-xl lg:col-span-2" />
        {Array.from({ length: 4 }, (_, i) => (
          <Skeleton key={i} className="h-[340px] rounded-xl" />
        ))}
      </div>
    </div>
  )
}
