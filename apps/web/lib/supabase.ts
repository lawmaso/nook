import { createClient } from "@supabase/supabase-js";

const supabaseURL = process.env.NEXT_PUBLIC_SUPABASE_URL as string
const supabaseKey = process.env.NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY as string

export const supabase = createClient(
    supabaseURL,
    supabaseKey
)
