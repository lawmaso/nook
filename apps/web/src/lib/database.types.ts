// Shape matches `supabase gen types typescript` output for the public schema.
// Regenerate after schema changes:
//   supabase gen types typescript --linked > apps/web/src/lib/database.types.ts
// (or `--local` with `supabase start` running)

export type Json =
  | string
  | number
  | boolean
  | null
  | { [key: string]: Json | undefined }
  | Json[]

export type Database = {
  public: {
    Tables: {
      scrape_entries: {
        Row: {
          acceptance_rate: number
          attempted: number
          avg_points: number
          avg_time_s: number
          last_active: string
          rank: number
          scrape_id: number
          solved: number
          total_points: number
          uid: number
        }
        Insert: {
          acceptance_rate: number
          attempted: number
          avg_points: number
          avg_time_s: number
          last_active: string
          rank: number
          scrape_id: number
          solved: number
          total_points: number
          uid: number
        }
        Update: {
          acceptance_rate?: number
          attempted?: number
          avg_points?: number
          avg_time_s?: number
          last_active?: string
          rank?: number
          scrape_id?: number
          solved?: number
          total_points?: number
          uid?: number
        }
        Relationships: [
          {
            foreignKeyName: "scrape_entries_scrape_id_fkey"
            columns: ["scrape_id"]
            isOneToOne: false
            referencedRelation: "scrapes"
            referencedColumns: ["scrape_id"]
          },
          {
            foreignKeyName: "scrape_entries_uid_fkey"
            columns: ["uid"]
            isOneToOne: false
            referencedRelation: "users"
            referencedColumns: ["uid"]
          },
        ]
      }
      scrape_failures: {
        Row: {
          error: string
          failed_at: string
          failure_id: number
          kind: Database["public"]["Enums"]["leaderboard"]
          persist_attempts: number
          scrape_attempts: number
          stage: Database["public"]["Enums"]["scrape_stage"]
        }
        Insert: {
          error: string
          failed_at: string
          failure_id?: never
          kind: Database["public"]["Enums"]["leaderboard"]
          persist_attempts: number
          scrape_attempts: number
          stage: Database["public"]["Enums"]["scrape_stage"]
        }
        Update: {
          error?: string
          failed_at?: string
          failure_id?: never
          kind?: Database["public"]["Enums"]["leaderboard"]
          persist_attempts?: number
          scrape_attempts?: number
          stage?: Database["public"]["Enums"]["scrape_stage"]
        }
        Relationships: []
      }
      scrapes: {
        Row: {
          duration_ms: number
          kind: Database["public"]["Enums"]["leaderboard"]
          period: string | null
          persist_attempts: number
          scrape_attempts: number
          scrape_id: number
          scraped_at: string
        }
        Insert: {
          duration_ms: number
          kind: Database["public"]["Enums"]["leaderboard"]
          period?: string | null
          persist_attempts?: number
          scrape_attempts?: number
          scrape_id?: never
          scraped_at: string
        }
        Update: {
          duration_ms?: number
          kind?: Database["public"]["Enums"]["leaderboard"]
          period?: string | null
          persist_attempts?: number
          scrape_attempts?: number
          scrape_id?: never
          scraped_at?: string
        }
        Relationships: []
      }
      users: {
        Row: {
          member_since: string | null
          uid: number
          username: string
        }
        Insert: {
          member_since?: string | null
          uid: number
          username: string
        }
        Update: {
          member_since?: string | null
          uid?: number
          username?: string
        }
        Relationships: []
      }
    }
    Views: {
      [_ in never]: never
    }
    Functions: {
      leaderboard_metrics: {
        Args: {
          p_bucket?: string
          p_from: string
          p_kind: Database["public"]["Enums"]["leaderboard"]
          p_to: string
        }
        Returns: {
          active_24h: number
          avg_acceptance: number
          avg_time_s: number
          bucket: string
          entrants: number
          median_points: number
          period: string | null
          scrape_id: number
          scraped_at: string
          top100_points: number | null
          top10_points: number | null
          top1_points: number | null
          top50_points: number | null
          total_solved: number
        }[]
      }
      user_history: {
        Args: {
          p_bucket?: string
          p_from: string
          p_kind: Database["public"]["Enums"]["leaderboard"]
          p_to: string
          p_uid: number
        }
        Returns: {
          acceptance_rate: number
          attempted: number
          avg_points: number
          avg_time_s: number
          bucket: string
          period: string | null
          rank: number
          scraped_at: string
          solved: number
          total_points: number
        }[]
      }
    }
    Enums: {
      leaderboard: "monthly" | "global"
      scrape_stage: "scrape" | "persist"
    }
    CompositeTypes: {
      [_ in never]: never
    }
  }
}

type PublicSchema = Database["public"]

export type Tables<T extends keyof PublicSchema["Tables"]> =
  PublicSchema["Tables"][T]["Row"]
export type Enums<T extends keyof PublicSchema["Enums"]> =
  PublicSchema["Enums"][T]
export type FunctionReturns<T extends keyof PublicSchema["Functions"]> =
  PublicSchema["Functions"][T]["Returns"]
