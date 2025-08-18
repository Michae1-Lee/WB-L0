package cache

import (
	"context"
	"github.com/pkg/errors"
	"wb/internal/repository"
)

type Warmer struct {
	cache     *Cache
	orderRepo *repository.OrderRepo
}

func NewWarmer(orderRepo *repository.OrderRepo, cache *Cache) *Warmer {
	return &Warmer{
		cache:     cache,
		orderRepo: orderRepo,
	}
}

func (w *Warmer) Warm(ctx context.Context) error {
	orders, err := w.orderRepo.GetLastOrders(ctx, w.cache.size)
	if err != nil {
		return errors.WithMessage(err, "warm: get last orders")
	}
	if len(orders) == 0 {
		return nil
	}
	for i := len(orders) - 1; i >= 0; i-- {
		w.cache.PutInCache(orders[i].OrderUID, orders[i])
	}
	return nil
}
