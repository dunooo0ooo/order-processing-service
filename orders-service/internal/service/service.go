package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dunooo0ooo/wb-tech-l0/orders-service/internal/entity"
	"github.com/dunooo0ooo/wb-tech-l0/orders-service/internal/infrastructure"
	"github.com/dunooo0ooo/wb-tech-l0/orders-service/internal/metrics"
	oc "github.com/dunooo0ooo/wb-tech-l0/orders-service/internal/order-cache"
	"go.uber.org/zap"
	"time"
)

var (
	ErrBadMessage = errors.New("bad message")
)

type Service struct {
	repo             infrastructure.Repository
	cache            oc.OrderCache
	logger           *zap.Logger
	met              *metrics.Metrics
	cacheWarmupLimit int
}

func New(repo infrastructure.Repository, cache oc.OrderCache, logger *zap.Logger, warmupLimit int, met *metrics.Metrics) *Service {
	if warmupLimit <= 0 {
		warmupLimit = 1000
	}
	return &Service{
		repo:             repo,
		cache:            cache,
		logger:           logger,
		met:              met,
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
		if s.met != nil {
			s.met.KafkaBad.Inc()
		}
		s.logger.Warn("bad message: json unmarshal", zap.Error(err))
		return fmt.Errorf("%w: %v", ErrBadMessage, err)
	}
	if o.OrderUID == "" {
		if s.met != nil {
			s.met.KafkaBad.Inc()
		}
		s.logger.Warn("bad message: empty order_uid")
		return fmt.Errorf("%w: empty order_uid", ErrBadMessage)
	}
	if o.DateCreated.IsZero() {
		o.DateCreated = time.Now().UTC()
	}
	start := time.Now()
	if err := s.repo.Save(ctx, &o); err != nil {
		if s.met != nil {
			s.met.DBSaveDuration.Observe(time.Since(start).Seconds())
			s.met.KafkaErrors.Inc()
		}
		s.logger.Error("save order failed", zap.String("order_uid", o.OrderUID), zap.Error(err))
		return err
	}

	if s.met != nil {
		s.met.DBSaveDuration.Observe(time.Since(start).Seconds())
		s.met.KafkaMessages.Inc()
	}

	s.cache.Set(o.OrderUID, &o)
	s.logger.Info("order saved", zap.String("order_uid", o.OrderUID))
	return nil
}

func (s *Service) GetOrder(ctx context.Context, id string) (*entity.Order, error) {
	if o, ok := s.cache.Get(id); ok {
		if s.met != nil {
			s.met.CacheHits.Inc()
		}
		return o, nil
	}
	if s.met != nil {
		s.met.CacheMisses.Inc()
	}
	s.logger.Debug("cache miss", zap.String("order_uid", id))

	start := time.Now()
	o, err := s.repo.GetByID(ctx, id)
	if s.met != nil {
		s.met.DBGetDuration.Observe(time.Since(start).Seconds())
	}
	if err != nil {
		s.logger.Error("get order from db failed", zap.String("order_uid", id), zap.Error(err))
		return nil, err
	}
	s.cache.Set(id, o)
	return o, nil
}
