import { createClient } from "@supabase/supabase-js"

import type { Database } from "@/lib/database.types"

const url = process.env.NEXT_PUBLIC_SUPABASE_URL
const key = process.env.NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY

if (!url || !key) {
  throw new Error(
    "Missing NEXT_PUBLIC_SUPABASE_URL or NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY (see apps/web/.env.example)",
  )
}

// No auth: every table is public-read for `anon` via RLS, so one stateless
// client is shared by all server-side reads.
export const supabase = createClient<Database>(url, key, {
  auth: { persistSession: false, autoRefreshToken: false },
})
