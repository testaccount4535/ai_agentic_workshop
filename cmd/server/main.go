// Command server runs the ride hailing API web server.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/testaccount4535/ai_agentic_workshop/internal/handler"
	"github.com/testaccount4535/ai_agentic_workshop/internal/store"
)

const (
	addr   = ":8080"
	dbPath = "rides.db"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	st, err := store.Open(dbPath, logger)
	if err != nil {
		logger.Error("startup failed: cannot open store", "error", err)
		os.Exit(1)
	}
	defer func() { _ = st.Close() }()

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler.New(st, logger).Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Shut down gracefully on interrupt/terminate.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		logger.Info("server listening", "addr", addr, "db_path", dbPath)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
	logger.Info("server stopped")
}
