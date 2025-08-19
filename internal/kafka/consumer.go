package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"time"
	"wb/internal/cache"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"wb/internal/models"
	"wb/internal/service"
)

type Consumer struct {
	reader  *kafka.Reader
	logger  *zap.SugaredLogger
	service *service.OrderService
	cache   *cache.Cache
}

func NewConsumer(brokers []string, topic, groupID string, logger *zap.SugaredLogger, svc *service.OrderService, cache *cache.Cache) *Consumer {
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		CommitInterval: 0,
		StartOffset:    kafka.FirstOffset,
	})

	return &Consumer{reader: r, logger: logger, service: svc, cache: cache}
}

func (c *Consumer) Start(ctx context.Context) error {
	cfg := c.reader.Config()
	c.logger.Infow("kafka consumer started",
		"brokers", cfg.Brokers, "topic", cfg.Topic, "group", cfg.GroupID)

	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				c.logger.Infow("kafka consumer stopped")
				return nil
			}
			c.logger.Errorw("fetch message failed", "err", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var order models.Order
		if err := json.Unmarshal(m.Value, &order); err != nil {
			c.logger.Errorw("bad message json", "err", err, "payload", string(m.Value))
			continue
		}

		if order.OrderUID == "" {
			c.logger.Errorw("message missing order_uid", "payload", string(m.Value))
			continue
		}

		orderUid, err := c.service.UpsertOrder(ctx, order)
		if err != nil {
			c.logger.Errorw("upsert failed", "order_uid", order.OrderUID, "err", err)
			continue
		}

		c.cache.PutInCache(orderUid, order)

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			c.logger.Errorw("commit failed", "order_uid", order.OrderUID, "err", err)
			continue
		}

		c.logger.Infow("message processed",
			"order_uid", order.OrderUID,
			"partition", m.Partition,
			"offset", m.Offset,
		)
	}
}
