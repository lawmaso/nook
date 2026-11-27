# Nook Project Notes

## Supabase

### Installation
```
brew install supabase/tap/supabase
supabase telemetry disable
supabase login
```

To create + apply new migration(s):
```
supabase migration new <migration_name>  # create migration
supabase migration up                    # apply migration
```

## Local Database
```sh
open -a Docker                  # ensure docker daemon is running
docker compose up -d --wait     # start local nookdb postgres instance
docker compose ps               # status should be "healthy"
docker compose logs -f nookdb   # tail logs
docker compose down             # stop container; keep data
docker compose down -v          # stop container; wipe data (volume deleted)
```

Init scripts mounted in `docker-entrypoint-initdb.d` run only on an empty volume. After changing the schema, run `down -v`, then `up -d`

```
export DATABASE_URL="postgres://nook:nook@localhost:5432/nookdb?sslmode=disable"
```

## Local API
```
export DATABASE_URL="postgres://nook:nook@localhost:5432/nookdb?sslmode=disable"
go run ./cmd/nookapi
curl localhost:8080/leaderboard/monthly | jq '.curr[:3]'
```

## Web App (`apps/web`)
Next.js (App Router, Cache Components) + shadcn/ui (base-nova, Base UI) reading Supabase's data API with the publishable key. Every table is public-read for `anon`, so there's no auth.

```sh
cd apps/web
cp .env.example .env.local      # fill in NEXT_PUBLIC_SUPABASE_URL / NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY
pnpm install
pnpm dev                        # http://localhost:3000
```

Routes:
- `/leaderboard/[monthly|global]`: latest snapshot, sortable table, head-to-head compare (`?a=<uid>&b=<uid>`)
- `/u/[uid]`: current standings + history charts (`?kind=&from=&to=`)
- `/analytics`: board-wide trends over a date range (`?kind=&from=&to=`, days in US Eastern)

Time-series charts call the `leaderboard_metrics` / `user_history` RPCs from `supabase/migrations/*_web_analytics_rpcs.sql`. They return the last scrape per hour/day/week (bucket chosen from the range length). Until that migration is applied, the chart sections show a "not deployed" notice:
```sh
supabase link --project-ref <ref>   # once
supabase db push                    # apply pending migrations to the hosted project
```

Regenerate types after schema changes:
```sh
supabase gen types typescript --linked > apps/web/src/lib/database.types.ts
```

Data is cached with the `scrape` cacheLife profile in `apps/web/next.config.ts` (revalidate every 5 min).
