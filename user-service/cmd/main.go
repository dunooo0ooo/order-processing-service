package main

import (
	"context"
	"github.com/dunooo0ooo/wb-tech-l0/user-service/internal/app"
	"github.com/dunooo0ooo/wb-tech-l0/user-service/internal/delivery"
	"github.com/dunooo0ooo/wb-tech-l0/user-service/internal/infrastructure/postgres"
	"github.com/dunooo0ooo/wb-tech-l0/user-service/internal/service"
	"github.com/dunooo0ooo/wb-tech-l0/user-service/pkg/config"
	"github.com/dunooo0ooo/wb-tech-l0/user-service/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	log, err := logger.New(cfg.Logger.Level)
	if err != nil {
		log.Fatal("failed to initialize logger", zap.Error(err))
	}
	defer func(log *zap.Logger) {
		_ = log.Sync()
	}(log)

	log.Info("config loaded",
		zap.String("http_addr", cfg.HTTP.Addr),
	)

	poolCfg, err := pgxpool.ParseConfig(cfg.Postgres.DSN())
	if err != nil {
		log.Fatal("cannot parse postgres DSN", zap.Error(err))
	}

	dbpool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		log.Fatal("cannot connect to postgres", zap.Error(err))
	}
	defer dbpool.Close()

	log.Info("connected to postgres",
		zap.String("host", cfg.Postgres.Host),
		zap.Int("port", cfg.Postgres.Port),
		zap.String("db", cfg.Postgres.DBName),
	)

	repository := postgres.NewUserRepository(dbpool)
	svc := service.NewUserService(repository, log)

	handler := delivery.New(svc)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	srv := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	application := app.NewApp(srv, log, &cfg)

	if err := application.Start(ctx); err != nil {
		log.Error("application error", zap.Error(err))
	}

	if err := application.Shutdown(); err != nil {
		log.Error("graceful shutdown failed", zap.Error(err))
	}
}
