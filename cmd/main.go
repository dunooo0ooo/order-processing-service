package main

import (
	"context"
	"errors"
	"github.com/dunooo0ooo/wb-tech-l0/internal/app"
	"github.com/dunooo0ooo/wb-tech-l0/internal/delivery"
	"github.com/dunooo0ooo/wb-tech-l0/internal/infrastructure/postgres"
	consumer "github.com/dunooo0ooo/wb-tech-l0/internal/kafka"
	"github.com/dunooo0ooo/wb-tech-l0/internal/metrics"
	oc "github.com/dunooo0ooo/wb-tech-l0/internal/order-cache"
	"github.com/dunooo0ooo/wb-tech-l0/internal/service"
	"github.com/dunooo0ooo/wb-tech-l0/pkg/config"
	"github.com/dunooo0ooo/wb-tech-l0/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
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

	repository := postgres.NewOrderRepository(dbpool)
	c := oc.NewOrderCache(cfg.Cache)

	reg := prometheus.NewRegistry()
	met := metrics.New(reg)

	svc := service.New(repository, c, log, cfg.Cache.Limit, met)

	if err := svc.WarmupCache(ctx); err != nil {
		log.Fatal("warmup cache failed", zap.Error(err))
	}

	cons := consumer.New(cfg.Kafka, svc, log)
	defer cons.Close()

	go func() {
		if err := cons.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Error("kafka consumer stopped", zap.Error(err))
			stop()
		}
	}()

	handler := delivery.NewOrderHandler(svc)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	mux.Handle("/", http.FileServer(http.Dir("./web")))
	mux.Handle("GET /metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

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
