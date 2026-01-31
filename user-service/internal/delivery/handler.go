package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dunooo0ooo/wb-tech-l0/user-service/internal/delivery/dto"
	"github.com/dunooo0ooo/wb-tech-l0/user-service/internal/entity"
	"github.com/dunooo0ooo/wb-tech-l0/user-service/internal/infrastructure"
	"github.com/dunooo0ooo/wb-tech-l0/user-service/internal/service"
	"net/http"
	"strings"
)

type UserService interface {
	TryCreate(ctx context.Context, username, password, role string) (bool, error)
	VerifyUser(ctx context.Context, username, password string) (entity.User, error)
}

type UserHandler struct {
	us UserService
}

func New(us UserService) *UserHandler {
	return &UserHandler{
		us: us,
	}
}

func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /users", h.CreateUser)
	mux.HandleFunc("GET /login", h.Verify)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Role = strings.TrimSpace(req.Role)

	if req.Role == "" {
		req.Role = "user"
	}
	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	ok, err := h.us.TryCreate(r.Context(), req.Username, req.Password, req.Role)
	if err != nil {
		if errors.Is(err, infrastructure.ErrUserAlreadyExist) {
			http.Error(w, "user already exist", http.StatusConflict)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(dto.CreateUserResponse{
		IsSuccess: ok,
	})
}

func (h *UserHandler) Verify(w http.ResponseWriter, r *http.Request) {
	var req dto.VerifyUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		http.Error(w, "invalid arguments", http.StatusBadRequest)
		return
	}

	u, err := h.us.VerifyUser(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidArguments) {
			http.Error(w, "invalid arguments", http.StatusBadRequest)
			return
		}
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dto.VerifyUserResponse{
		Role: string(u.Role),
	})
}
