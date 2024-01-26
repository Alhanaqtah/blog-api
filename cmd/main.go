package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"blog-api/internal/config"
	"blog-api/internal/lib/logger"
	"blog-api/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.MustLoad()

	log := logger.New(cfg.Env)

	log.Debug("initializing server...", slog.String("addr", cfg.Address))

	// <- mux and middleware
	r := chi.NewRouter()

	srv := http.Server{
		Handler:      r,
		Addr:         cfg.Address,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	log.Debug("server initialized")
	log.Info("server is running...")

	// Gracefully shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("error starting sever", sl.Error(err))
		}
	}()

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	srv.Shutdown(ctx)

	log.Info("server stopped")
}
