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
)

type nookServer struct {
	cache *cache.LeaderboardCache
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

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	lbCache := cache.New(scraper.NewClient())
	lbCache.OnRefreshed = func(lb models.Leaderboard, curr []*models.User) {
		slog.Info("refreshed", "leaderboard", lb.DBKind())
	}

	interval := cache.ParseRefreshInterval()
	lbCache.LaunchRefresher(ctx, interval)

	ns := &nookServer{cache: lbCache}

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
