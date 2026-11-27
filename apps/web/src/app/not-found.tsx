import Link from "next/link"

import { buttonVariants } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

export default function NotFound() {
  return (
    <Card className="mx-auto mt-12 max-w-lg">
      <CardHeader>
        <CardTitle>Not found</CardTitle>
        <CardDescription>That leaderboard or user doesn&apos;t exist.</CardDescription>
      </CardHeader>
      <CardContent>
        <Link href="/leaderboard/monthly" className={buttonVariants()}>
          Back to the leaderboard
        </Link>
      </CardContent>
    </Card>
  )
}
