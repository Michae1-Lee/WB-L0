package main

import (
	"context"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"wb/internal/cache"
	"wb/internal/handler"
	"wb/internal/kafka"
	"wb/internal/repository"
	"wb/internal/service"
	ui2 "wb/internal/ui"
	"wb/pkg/postgres"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()
	cacheSize := 15

	ctx := context.Background()

	db, err := postgres.InitDB(sugar)
	if err != nil {
		sugar.Fatalf("Init db failed: %v", err)
	}

	orderCache := cache.NewCache(cacheSize)
	deliveryRepo := repository.NewDeliveryRepo(db)
	itemRepo := repository.NewItemRepo(db)
	paymentRepo := repository.NewPaymentRepo(db)
	orderRepo := repository.NewOrderRepo(db, deliveryRepo, paymentRepo, itemRepo)
	orderService := service.NewOrderService(sugar, orderRepo)
	orderHandler := handler.NewOrderHandler(sugar, orderService, orderCache)
	ui := ui2.NewSimpleUi()
	warmer := cache.NewWarmer(orderRepo, orderCache)

	err = warmer.Warm(ctx)
	if err != nil {
		return
	}

	orderCache.Show()

	httpAddr := os.Getenv("HTTP_ADDR")
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	topic := os.Getenv("KAFKA_TOPIC")
	group := os.Getenv("KAFKA_GROUP")

	consumer := kafka.NewConsumer(brokers, topic, group, sugar, orderService, orderCache)

	go func() {
		if err := consumer.Start(ctx); err != nil {
			sugar.Errorw("kafka consumer stopped with error", "err", err)
		}
	}()

	producer := kafka.NewProducer(brokers, topic, sugar)

	go func() {
		for {
			if err := producer.Run(ctx, time.Second*15); err != nil {
				sugar.Errorw("failed to produce order", "err", err)
			}
		}
	}()

	orderRouter := http.NewServeMux()
	orderRouter.HandleFunc("GET /order/{order_uid}", orderHandler.GetOrder)
	orderRouter.HandleFunc("GET /", ui.Index)
	server := http.Server{
		Addr:    httpAddr,
		Handler: orderRouter,
	}

	log.Println("server started")
	server.ListenAndServe()
}
