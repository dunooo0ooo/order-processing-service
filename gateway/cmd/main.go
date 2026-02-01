package main

import (
	"context"
	"errors"
	"github.com/dunooo0ooo/wb-tech-l0/gateway/internal/delivery"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/dunooo0ooo/wb-tech-l0/gateway/internal/clients"
	"github.com/dunooo0ooo/wb-tech-l0/gateway/internal/proxy"
	"github.com/dunooo0ooo/wb-tech-l0/gateway/pkg/config"
)

func main() {
	conf := config.NewConfig()

	userClient := clients.NewUserClient(conf.UserSvc)
	ordersProxy := proxy.New(conf.OrdersSvc) // http.Handler

	mux := http.NewServeMux()
	h := delivery.New(userClient, ordersProxy, conf.Secret, 24*time.Hour)
	h.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:         conf.Addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("gateway listening on %s", conf.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen error: %v", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
		_ = srv.Close()
	}

	log.Printf("gateway stopped")
}
