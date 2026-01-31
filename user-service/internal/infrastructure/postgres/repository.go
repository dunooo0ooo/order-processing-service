package postgres

import (
	"context"
	"errors"
	"github.com/dunooo0ooo/wb-tech-l0/user-service/internal/entity"
	"github.com/dunooo0ooo/wb-tech-l0/user-service/internal/infrastructure"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (repo *Repository) TryCreate(ctx context.Context, user *entity.User) (bool, error) {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return false, infrastructure.ErrInternalDatabase
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, `
			INSERT INTO users(username, password, role, created_at)
			VALUES(@username, @password, @role, now())`,
		pgx.NamedArgs{
			"username": user.Username,
			"password": user.Password,
			"role":     user.Role})

	if err != nil {
		return false, infrastructure.ErrInternalDatabase
	}

	err = tx.Commit(ctx)
	if err != nil {
		return false, infrastructure.ErrInternalDatabase
	}
	return true, nil
}

func (repo *Repository) GetUserByUsername(ctx context.Context, username string) (entity.User, error) {
	var u entity.User

	err := repo.pool.QueryRow(ctx, `
		SELECT username, password, role, created_at
		FROM users
		WHERE username = @username
	`, pgx.NamedArgs{"username": username}).Scan(
		&u.Username,
		&u.Password, // хэш
		&u.Role,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, infrastructure.ErrUserNotFound
		}
		return entity.User{}, infrastructure.ErrInternalDatabase
	}

	return u, nil
}
