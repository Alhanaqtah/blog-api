package main

import (
	"blog-api/internal/http-server/handlers/user"
	"context"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"blog-api/internal/config"
	"blog-api/internal/lib/logger"
	"blog-api/internal/lib/logger/sl"
	userservice "blog-api/internal/service/user"
	"blog-api/internal/storage/sqlite"

	_ "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.MustLoad()

	log := logger.New(cfg.Env)

	log.Debug("initializing server...", slog.String("addr", cfg.Address))

	// Init storage
	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("error opening storage")
		return
	}

	// Init service layer
	usrService := userservice.New(log, storage, cfg.TokenTTL)

	// Handlers and middleware
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Init handlers
	usr := user.New(log, usrService, cfg.Secret)

	r.Route("/users", usr.Register())
	//r.Route("/article", art.Register())

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
