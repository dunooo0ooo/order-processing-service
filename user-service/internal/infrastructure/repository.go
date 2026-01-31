package infrastructure

import (
	"context"
	"errors"
	"github.com/dunooo0ooo/wb-tech-l0/user-service/internal/entity"
)

type UserRepository interface {
	TryCreate(ctx context.Context, user *entity.User) (bool, error)
	GetUserByUsername(ctx context.Context, username string) (entity.User, error)
}

var (
	ErrInternalDatabase = errors.New("internal database error")
	ErrUserNotFound     = errors.New("user not found")
)
