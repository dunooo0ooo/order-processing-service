package delivery

import (
	"context"
	"encoding/json"
	"github.com/dunooo0ooo/wb-tech-l0/internal/entity"
	"net/http"
)

type OrderService interface {
	SaveOrderFromEvent(ctx context.Context, msg []byte) error
	GetOrder(ctx context.Context, id string) (*entity.Order, error)
}

type OrderHandler struct {
	os OrderService
}

func NewOrderHandler(os OrderService) *OrderHandler {
	return &OrderHandler{os: os}
}

func (h *OrderHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /order/{id}", h.GetOrderInfo)
}

func (h *OrderHandler) GetOrderInfo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	o, err := h.os.GetOrder(r.Context(), id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(o)
}
