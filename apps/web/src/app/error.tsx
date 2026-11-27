"use client"

import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

export default function ErrorPage({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) {
  return (
    <Card className="mx-auto mt-12 max-w-lg">
      <CardHeader>
        <CardTitle>Something went wrong</CardTitle>
        <CardDescription>{error.message || "Failed to load data from Supabase."}</CardDescription>
      </CardHeader>
      <CardContent>
        <Button onClick={reset}>Try again</Button>
      </CardContent>
    </Card>
  )
}
