package consumer

import (
	"context"
	"errors"
	"github.com/dunooo0ooo/wb-tech-l0/orders-service/internal/delivery"
	"github.com/dunooo0ooo/wb-tech-l0/orders-service/internal/service"
	"github.com/dunooo0ooo/wb-tech-l0/orders-service/pkg/config"
	"go.uber.org/zap"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	r   *kafka.Reader
	svc delivery.OrderService
	log *zap.Logger
}

func New(cfg config.KafkaConsumerConfig, svc delivery.OrderService, logger *zap.Logger) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		MinBytes:       1e3,
		MaxBytes:       10e6,
		CommitInterval: 0,
	})

	return &Consumer{r: r, svc: svc, log: logger}
}

func (c *Consumer) Close() error { return c.r.Close() }

func (c *Consumer) Run(ctx context.Context) error {
	backoff := 200 * time.Millisecond

	for {
		m, err := c.r.FetchMessage(ctx)
		if err != nil {
			return err
		}

		switch m.Topic {
		case "orders_creation":
			err := c.svc.SaveOrderFromEvent(ctx, m.Value)
			if err != nil {
				c.log.Warn("failed to save order", zap.Error(err))
			}
		}

		if errors.Is(err, service.ErrBadMessage) {
			c.log.Fatal("bad message, skip: %v", zap.Error(err))
			if e := c.r.CommitMessages(ctx, m); e != nil {
				c.log.Fatal("commit error: %v", zap.Error(e))
			}
			continue
		}

		if err != nil {
			c.log.Fatal("handle error (no commit): %v", zap.Error(err))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			if backoff < 5*time.Second {
				backoff *= 2
			}
			continue
		}

		if err := c.r.CommitMessages(ctx, m); err != nil {
			c.log.Fatal("commit error: %v", zap.Error(err))
		}

		backoff = 200 * time.Millisecond
	}
}
