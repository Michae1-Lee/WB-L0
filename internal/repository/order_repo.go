package repository

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"wb/internal/models"
)

type OrderRepo struct {
	db         *sqlx.DB
	deliveries *DeliveryRepo
	payments   *PaymentRepo
	items      *ItemRepo
}

func NewOrderRepo(db *sqlx.DB, d *DeliveryRepo, p *PaymentRepo, i *ItemRepo) *OrderRepo {
	return &OrderRepo{
		db:         db,
		deliveries: d,
		payments:   p,
		items:      i,
	}
}

func (r *OrderRepo) Upsert(ctx context.Context, o models.Order) (order_uid string, err error) {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return "", errors.WithMessage(err, "begin tx")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	const upsertOrder = `
		INSERT INTO orders (
			order_uid, track_number, entry, locale, internal_signature, customer_id,
			delivery_service, shardkey, sm_id, date_created, oof_shard
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11
		)
		ON CONFLICT (order_uid) DO UPDATE SET
			track_number=$2,
			entry=$3,
			locale=$4,
			internal_signature=$5,
			customer_id=$6,
			delivery_service=$7,
			shardkey=$8,
			sm_id=$9,
			date_created=$10,
			oof_shard=$11
	`
	if _, err = tx.ExecContext(ctx, upsertOrder,
		o.OrderUID, o.TrackNumber, o.Entry, o.Locale, o.InternalSignature, o.CustomerId,
		o.DeliveryService, o.ShardKey, o.SmId, o.DateCreated, o.OofShard,
	); err != nil {
		return "", errors.WithMessage(err, "upsert orders")
	}

	o.Delivery.OrderUID = o.OrderUID
	if _, err = r.deliveries.UpsertTx(ctx, tx, o.Delivery); err != nil {
		return "", errors.WithMessage(err, "upsert delivery")
	}

	o.Payment.OrderUID = o.OrderUID
	if _, err = r.payments.UpsertTx(ctx, tx, o.Payment); err != nil {
		return "", errors.WithMessage(err, "upsert payment")
	}

	if _, err = r.items.DeleteByOrderUIDTx(ctx, tx, o.OrderUID); err != nil {
		return "", errors.WithMessage(err, "delete items by order_uid")
	}

	for _, it := range o.Items {
		it.OrderUID = o.OrderUID
		if it.TrackNumber == "" {
			it.TrackNumber = o.TrackNumber
		}
		if _, err = r.items.UpsertTx(ctx, tx, it); err != nil {
			return "", errors.WithMessage(err, "upsert item")
		}
	}

	if err = tx.Commit(); err != nil {
		return "", errors.WithMessage(err, "commit")
	}
	return o.OrderUID, nil
}

func (r *OrderRepo) Get(ctx context.Context, orderUID string) (*models.Order, error) {
	var o models.Order

	const selOrder = `
		SELECT order_uid, track_number, entry, locale, internal_signature, customer_id,
		       delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders
		WHERE order_uid = $1
	`
	if err := r.db.GetContext(ctx, &o, selOrder, orderUID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.WithMessage(err, "select order")
	}

	if d, err := r.deliveries.Get(ctx, models.Delivery{OrderUID: orderUID}); err != nil {
		return nil, err
	} else if d != nil {
		o.Delivery = *d
	}

	if p, err := r.payments.Get(ctx, orderUID); err != nil {
		return nil, err
	} else if p != nil {
		o.Payment = *p
	}

	items, err := r.items.GetByOrderUID(ctx, orderUID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return &o, nil
}

func (r *OrderRepo) Delete(ctx context.Context, orderUID string) (int64, error) {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, errors.WithMessage(err, "begin tx")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	res, err := tx.ExecContext(ctx, `DELETE FROM orders WHERE order_uid=$1`, orderUID)
	if err != nil {
		return 0, errors.WithMessage(err, "delete order")
	}
	aff, _ := res.RowsAffected()

	if err = tx.Commit(); err != nil {
		return 0, errors.WithMessage(err, "commit")
	}
	return aff, nil
}

func (r *OrderRepo) GetLastOrders(ctx context.Context, n int) ([]models.Order, error) {
	const selOrders = `
		SELECT order_uid, track_number, entry, locale, internal_signature, customer_id,
		       delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders
		ORDER BY date_created DESC, order_uid DESC
		LIMIT $1
	`

	var orders []models.Order
	if err := r.db.SelectContext(ctx, &orders, selOrders, n); err != nil {
		return nil, errors.WithMessage(err, "select last orders")
	}
	if len(orders) == 0 {
		return nil, nil
	}

	for i := range orders {
		uid := orders[i].OrderUID

		if d, err := r.deliveries.Get(ctx, models.Delivery{OrderUID: uid}); err != nil {
			return nil, err
		} else if d != nil {
			orders[i].Delivery = *d
		}

		if p, err := r.payments.Get(ctx, uid); err != nil {
			return nil, err
		} else if p != nil {
			orders[i].Payment = *p
		}

		items, err := r.items.GetByOrderUID(ctx, uid)
		if err != nil {
			return nil, err
		}
		orders[i].Items = items
	}

	return orders, nil
}
