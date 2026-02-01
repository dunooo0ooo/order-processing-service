package delivery

import (
	"bytes"
	"encoding/json"
	"github.com/dunooo0ooo/wb-tech-l0/gateway/internal/delivery/dto"
	tok "github.com/dunooo0ooo/wb-tech-l0/gateway/internal/jwt"
	"io"
	"net/http"
	"time"
)

type UserClient interface {
	Register(r *http.Request) (*http.Response, error)
	Verify(username, password string) (role string, err error)
}

type OrdersProxy interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type Handler struct {
	uc     UserClient
	orders OrdersProxy
	secret string
	ttl    time.Duration
}

func New(uc UserClient, orders OrdersProxy, secret string, ttl time.Duration) *Handler {
	return &Handler{uc: uc, orders: orders, secret: secret, ttl: ttl}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	role, err := h.uc.Verify(req.Username, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := tok.Sign(h.secret, role, h.ttl)
	if err != nil {
		http.Error(w, "cannot sign token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dto.LoginResp{AccessToken: token})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	req, err := http.NewRequest(http.MethodPost, "", &buf)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.uc.Register(req)
	if err != nil {
		http.Error(w, "user service unavailable", http.StatusBadGateway)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
