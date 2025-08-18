package repository

import (
	"context"
	"database/sql"
	"wb/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type PaymentRepo struct {
	db *sqlx.DB
}

func NewPaymentRepo(db *sqlx.DB) *PaymentRepo {
	return &PaymentRepo{db: db}
}

type Payment struct {
	OrderUID     string `json:"order_uid" db:"order_uid"`
	Transaction  string `json:"transaction" db:"transaction"`
	RequestId    string `json:"request_id" db:"request_id"`
	Currency     string `json:"currency" db:"currency"`
	Provider     string `json:"provider" db:"provider"`
	Amount       int    `json:"amount" db:"amount"`
	PaymentDt    int64  `json:"payment_dt" db:"payment_dt"`
	Bank         string `json:"bank" db:"bank"`
	DeliveryCost int    `json:"delivery_cost" db:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total" db:"goods_total"`
	CustomFee    int    `json:"custom_fee" db:"custom_fee"`
}

func (r *PaymentRepo) UpsertTx(ctx context.Context, tx *sqlx.Tx, p models.Payment) (string, error) {
	const q = `
		INSERT INTO payments (
			order_uid, transaction, request_id, currency, provider, amount,
			payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11
		)
		ON CONFLICT (order_uid) DO UPDATE SET
			transaction   = EXCLUDED.transaction,
			request_id    = EXCLUDED.request_id,
			currency      = EXCLUDED.currency,
			provider      = EXCLUDED.provider,
			amount        = EXCLUDED.amount,
			payment_dt    = EXCLUDED.payment_dt,
			bank          = EXCLUDED.bank,
			delivery_cost = EXCLUDED.delivery_cost,
			goods_total   = EXCLUDED.goods_total,
			custom_fee    = EXCLUDED.custom_fee
		RETURNING order_uid
	`
	var uid string
	if err := tx.QueryRowxContext(ctx, q,
		p.OrderUID, p.Transaction, p.RequestId, p.Currency, p.Provider, p.Amount,
		p.PaymentDt, p.Bank, p.DeliveryCost, p.GoodsTotal, p.CustomFee,
	).Scan(&uid); err != nil {
		return "", errors.WithMessage(err, "upsert payment (tx)")
	}
	return uid, nil
}

func (r *PaymentRepo) Get(ctx context.Context, orderUID string) (*models.Payment, error) {
	const q = `
		SELECT order_uid, transaction, request_id, currency, provider, amount,
		       payment_dt, bank, delivery_cost, goods_total, custom_fee
		FROM payments
		WHERE order_uid = $1
	`
	var out models.Payment
	if err := r.db.GetContext(ctx, &out, q, orderUID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.WithMessage(err, "select payment by order_uid")
	}
	return &out, nil
}
