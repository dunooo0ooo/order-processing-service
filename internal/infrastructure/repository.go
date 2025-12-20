package infrastructure

import (
	"context"
	"errors"
	"github.com/dunooo0ooo/wb-tech-l0/internal/entity"
)

type Repository interface {
	Save(ctx context.Context, o *entity.Order) error
	GetByID(ctx context.Context, id string) (*entity.Order, error)
	LoadRecent(ctx context.Context, limit int) ([]entity.Order, error)
}

var (
	ErrInternalDatabase = errors.New("internal database error")
	ErrOrderNotFound    = errors.New("order not found")
)
