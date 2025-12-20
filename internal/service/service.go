package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dunooo0ooo/wb-tech-l0/internal/infrastructure"
	oc "github.com/dunooo0ooo/wb-tech-l0/internal/order-cache"
	"go.uber.org/zap"
	"time"

	"github.com/dunooo0ooo/wb-tech-l0/internal/entity"
)

var (
	ErrBadMessage = errors.New("bad message")
)

type Service struct {
	repo   infrastructure.Repository
	cache  oc.OrderCache
	logger *zap.Logger

	cacheWarmupLimit int
}

func New(repo infrastructure.Repository, cache oc.OrderCache, logger *zap.Logger, warmupLimit int) *Service {
	if warmupLimit <= 0 {
		warmupLimit = 1000
	}
	return &Service{
		repo:             repo,
		cache:            cache,
		logger:           logger,
		cacheWarmupLimit: warmupLimit,
	}
}

func (s *Service) WarmupCache(ctx context.Context) error {
	orders, err := s.repo.LoadRecent(ctx, s.cacheWarmupLimit)
	if err != nil {
		s.logger.Error("warmup cache failed", zap.Error(err))
		return err
	}
	for i := range orders {
		o := orders[i]
		s.cache.Set(o.OrderUID, &o)
	}
	s.logger.Info("cache warmed", zap.Int("count", len(orders)))
	return nil
}

func (s *Service) SaveOrderFromEvent(ctx context.Context, msg []byte) error {
	var o entity.Order
	if err := json.Unmarshal(msg, &o); err != nil {
		s.logger.Warn("bad message: json unmarshal", zap.Error(err))
		return fmt.Errorf("%w: %v", ErrBadMessage, err)
	}
	if o.OrderUID == "" {
		s.logger.Warn("bad message: empty order_uid")
		return fmt.Errorf("%w: empty order_uid", ErrBadMessage)
	}
	if o.DateCreated.IsZero() {
		o.DateCreated = time.Now().UTC()
	}

	if err := s.repo.Save(ctx, &o); err != nil {
		s.logger.Error("save order failed", zap.String("order_uid", o.OrderUID), zap.Error(err))
		return err
	}

	s.cache.Set(o.OrderUID, &o)
	s.logger.Info("order saved", zap.String("order_uid", o.OrderUID))
	return nil
}

func (s *Service) GetOrder(ctx context.Context, id string) (*entity.Order, error) {
	if o, ok := s.cache.Get(id); ok {
		s.logger.Debug("cache hit", zap.String("order_uid", id))
		return o, nil
	}
	s.logger.Debug("cache miss", zap.String("order_uid", id))

	o, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("get order from db failed", zap.String("order_uid", id), zap.Error(err))
		return nil, err
	}
	s.cache.Set(id, o)
	return o, nil
}
