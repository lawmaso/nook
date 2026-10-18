# Nook Project Notes

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
