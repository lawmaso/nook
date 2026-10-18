package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"nook/internal/cache"
	"nook/internal/models"
	"nook/internal/scraper"
	"nook/internal/store"

	"github.com/jackc/pgx/v5/pgxpool"
)

type nookServer struct {
	cache *cache.LeaderboardCache
	store *store.Store
}

type GetLeaderboardResponse struct {
	Prev []*models.User `json:"prev"`
	Curr []*models.User `json:"curr"`
}

func (ns *nookServer) handleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	kind, err := models.ParseLeaderboard(r.PathValue("kind"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cache, err := ns.cache.GetCachedLeaderboard(kind)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetLeaderboardResponse{
		Prev: cache.Prev,
		Curr: cache.Curr,
	})
}

// Loads the latest two scrapes per leaderboard.
// Allows a restart to immediately serve real data instead of
// waiting for the next interval to refresh.
func warmCache(ctx context.Context, lbCache *cache.LeaderboardCache, st *store.Store) {
	for _, lb := range []models.Leaderboard{models.MonthlyLeaderboard, models.GlobalLeaderboard} {
		curr, prev, err := st.LatestTwo(ctx, lb)
		if err != nil {
			slog.Error("failed to warm cache from db", "leaderboard", lb.String(), "error", err)
			continue
		}

		lbCache.Seed(lb, curr, prev)
	}
}

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	db := os.Getenv("DATABASE_URL")

	pool, err := pgxpool.New(ctx, db)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	st := store.New(pool)
	lbCache := cache.New(scraper.NewClient())

	lbCache.OnRefreshed = func(lb models.Leaderboard, curr []*models.User) {
		if err := st.SaveScrape(ctx, lb, curr); err != nil {
			// Peristent failure shouldn't affect live serving
			slog.Error("failed to persist scrape", "leaderboard", lb.String(), "error", err)
		} else {
			slog.Info("refreshed successfully", "leaderbaord", lb.String())
		}
	}

	warmCache(ctx, lbCache, st)

	interval := cache.ParseRefreshInterval()
	lbCache.LaunchRefresher(ctx, interval)

	ns := &nookServer{cache: lbCache, store: st}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /leaderboard/{kind}", ns.handleGetLeaderboard)

	httpServer := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		slog.Info("nook server listening", "addr", httpServer.Addr, "refresh_interval", interval)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	<-ctx.Done() // block until ctrl-c or SIGTERM
	slog.Info("shutting down nook server")
	httpServer.Shutdown(context.Background())
}
