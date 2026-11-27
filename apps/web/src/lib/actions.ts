"use server"

import { searchUsers, type UserHit } from "@/lib/queries"

export async function searchUsersAction(query: string): Promise<UserHit[]> {
  if (typeof query !== "string" || query.length > 64) return []
  return searchUsers(query)
}
