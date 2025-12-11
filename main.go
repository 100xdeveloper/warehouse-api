package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"warehouse-api/database"
	"warehouse-api/internal/api/handlers"
	"warehouse-api/internal/api/router"
	"warehouse-api/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// 1. Setup Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 2. Load Config
	cfg := config.Load()

	// 3. Database Connection
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		logger.Error("Unable to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// 4. Initialize Components
	store := database.New(pool)
	productHandler := handlers.NewProductHandler(store)
	r := router.New(productHandler)

	// 5. Server Setup with Graceful Shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Run server in a goroutine
	go func() {
		logger.Info("Server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Server shutting down...")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}
