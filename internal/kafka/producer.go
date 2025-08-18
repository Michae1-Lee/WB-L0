package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
	"wb/internal/models"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer struct {
	w      *kafka.Writer
	topic  string
	logger *zap.SugaredLogger
}

func NewProducer(brokers []string, topic string, logger *zap.SugaredLogger) *Producer {
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}
	return &Producer{w: w, topic: topic, logger: logger}
}

func (p *Producer) Close() error { return p.w.Close() }

func (p *Producer) Run(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	p.logger.Infow("kafka producer started", "topic", p.topic)

	for {
		select {
		case <-ctx.Done():
			p.logger.Infow("kafka producer stopped")
			return nil
		case <-ticker.C:
			key, val, orderUID := randomOrderJSON()
			err := p.w.WriteMessages(ctx, kafka.Message{
				Key:   []byte(key),
				Value: val,
				Time:  time.Now(),
			})
			if err != nil {
				p.logger.Errorw("produce failed", "order_uid", orderUID, "err", err)
				continue
			}
			p.logger.Infow("produced order", "order_uid", orderUID)
		}
	}
}

func randomOrderJSON() (key string, value []byte, orderUID string) {
	now := time.Now().UTC()
	orderUID = fmt.Sprintf("ord-%d", rand.Int63())
	track := fmt.Sprintf("TRK%06d", rand.Intn(1_000_000))
	o := models.Order{
		OrderUID:    orderUID,
		TrackNumber: track,
		Entry:       "WBIL",
		Delivery: models.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: models.Payment{
			Transaction:  orderUID,
			RequestId:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1000 + rand.Intn(2000),
			PaymentDt:    now.Unix(),
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtId:      900000 + rand.Intn(999999),
				TrackNumber: track,
				Price:       300 + rand.Intn(700),
				Rid:         "rid" + string(rune(rand.Intn(9999))),
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmId:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerId:        "test",
		DeliveryService:   "meest",
		ShardKey:          "9",
		SmId:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}

	b, _ := json.Marshal(o)
	return orderUID, b, orderUID
}
