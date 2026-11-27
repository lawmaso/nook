import { DatabaseIcon } from "lucide-react"

import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

/** Shown when the analytics RPCs (supabase/migrations/*_web_analytics_rpcs.sql) aren't deployed. */
export function MigrationNotice() {
  return (
    <Card className="border-dashed">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <DatabaseIcon className="size-4" /> Analytics functions not deployed
        </CardTitle>
        <CardDescription>
          This Supabase project doesn&apos;t have <code>leaderboard_metrics</code> / <code>user_history</code> yet. Apply
          the <code>web_analytics_rpcs</code> migration with <code>supabase db push</code>, then reload.
        </CardDescription>
      </CardHeader>
    </Card>
  )
}
