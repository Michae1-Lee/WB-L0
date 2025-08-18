package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"net/http"
	"wb/internal/cache"
	"wb/internal/dto"
	"wb/internal/models"
)

type OrderService interface {
	GetOrder(ctx context.Context, orderUID string) (*models.Order, error)
}

type OrderHandler struct {
	service OrderService
	logger  *zap.SugaredLogger
	cache   *cache.Cache
}

func NewOrderHandler(logger *zap.SugaredLogger, service OrderService, cache *cache.Cache) *OrderHandler {
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}
	return &OrderHandler{
		service: service,
		logger:  logger,
		cache:   cache,
	}
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderUID := r.PathValue("order_uid")
	if orderUID == "" {
		http.Error(w, "missing order_uid", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	if cachedOrder, ok := h.cache.GetIfInCache(orderUID); ok {

		var resp dto.OrderResponse
		if err := copier.Copy(&resp, &cachedOrder); err != nil {
			h.logger.Errorw("dto mapping failed (cache)", "order_uid", orderUID, "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		fmt.Println("From cache")
		writeJSON(w, resp)
		return
	}

	order, err := h.service.GetOrder(ctx, orderUID)

	if err != nil {
		h.logger.Errorw("get order failed", "order_uid", orderUID, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if order == nil {
		http.NotFound(w, r)
		return
	}

	h.cache.PutInCache(orderUID, *order)

	var resp dto.OrderResponse
	if err := copier.Copy(&resp, order); err != nil {
		h.logger.Errorw("dto mapping failed", "order_uid", orderUID, "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	fmt.Println("From db")
	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
