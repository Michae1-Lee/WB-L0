package service

import (
	"context"
	"go.uber.org/zap"
	"wb/internal/models"
)

type OrderRepo interface {
	Get(ctx context.Context, orderUID string) (*models.Order, error)
	Upsert(ctx context.Context, order models.Order) (string, error)
	Delete(ctx context.Context, orderUID string) (int64, error)
}

type OrderService struct {
	orderRepo OrderRepo
	logger    *zap.SugaredLogger
}

func NewOrderService(logger *zap.SugaredLogger, orderRepo OrderRepo) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		logger:    logger,
	}
}

func (s *OrderService) GetOrder(ctx context.Context, orderUID string) (*models.Order, error) {
	return s.orderRepo.Get(ctx, orderUID)
}

func (s *OrderService) UpsertOrder(ctx context.Context, order models.Order) (orderUid string, err error) {
	orderUid, err = s.orderRepo.Upsert(ctx, order)
	if err != nil {
		return orderUid, err
	}
	return orderUid, nil
}
