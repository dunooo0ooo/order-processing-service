package service

import (
	"context"
	"errors"
	"github.com/dunooo0ooo/wb-tech-l0/user-service/internal/entity"
	"github.com/dunooo0ooo/wb-tech-l0/user-service/internal/infrastructure"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo infrastructure.UserRepository
	log  *zap.Logger
}

func NewUserService(repo infrastructure.UserRepository, logger *zap.Logger) *UserService {
	return &UserService{repo: repo, log: logger}
}

const EmptyString = ""

var (
	ErrInvalidArguments = errors.New("invalid arguments")
)

func (u UserService) TryCreate(ctx context.Context, username, password, role string) (bool, error) {
	if username == EmptyString || password == EmptyString || role == EmptyString {
		u.log.Error("empty arguments")
		return false, ErrInvalidArguments
	}

	pass, err := hashPassword(password)
	if err != nil {
		u.log.Error("failed to hash password", zap.String("password", password))
		return false, err
	}

	user := entity.User{
		Username: username,
		Password: pass,
		Role:     entity.UserRole(role),
	}

	res, err := u.repo.TryCreate(ctx, &user)
	if err != nil {
		u.log.Error("failed to try create user", zap.Error(err))
		return false, err
	}

	u.log.Info("successfully created user", zap.String("username", username))
	return res, nil
}

func (u UserService) VerifyUser(ctx context.Context, username, password string) (entity.User, error) {
	if username == EmptyString || password == EmptyString {
		u.log.Error("empty arguments")
		return entity.User{}, ErrInvalidArguments
	}

	user, err := u.repo.GetUserByUsername(ctx, username)
	if err != nil {
		u.log.Error("failed to get user by username", zap.String("username", username))
		return entity.User{}, err
	}

	res := checkPassword(user.Password, password)
	if !res {
		u.log.Error("invalid password", zap.String("username", username))
		return entity.User{}, ErrInvalidArguments
	}
	u.log.Info("successfully verified user", zap.String("username", username))
	return user, nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func checkPassword(hash string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
