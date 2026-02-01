package delivery

import (
	"net/http"

	mw "github.com/dunooo0ooo/wb-tech-l0/gateway/internal/middleware"
)

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/login", h.Login)
	mux.HandleFunc("POST /auth/register", h.Register)

	mux.Handle("/order/", mw.Auth(h.secret, "user", "admin")(http.HandlerFunc(h.orders.ServeHTTP)))
	mux.Handle("/health", mw.Auth(h.secret, "admin")(http.HandlerFunc(h.orders.ServeHTTP)))
	mux.Handle("/metrics", mw.Auth(h.secret, "admin")(http.HandlerFunc(h.orders.ServeHTTP)))
}
