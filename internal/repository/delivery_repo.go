package repository

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"wb/internal/models"
)

type DeliveryRepo struct {
	db *sqlx.DB
}

func NewDeliveryRepo(db *sqlx.DB) *DeliveryRepo {
	return &DeliveryRepo{
		db: db,
	}
}

func (r *DeliveryRepo) UpsertTx(ctx context.Context, tx *sqlx.Tx, d models.Delivery) (string, error) {
	const q = `
		INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (order_uid) DO UPDATE SET
		  name=EXCLUDED.name,
		  phone=EXCLUDED.phone,
		  zip=EXCLUDED.zip,
		  city=EXCLUDED.city,
		  address=EXCLUDED.address,
		  region=EXCLUDED.region,
		  email=EXCLUDED.email
		RETURNING order_uid
	`
	var uid string
	if err := tx.QueryRowxContext(ctx, q,
		d.OrderUID, d.Name, d.Phone, d.Zip, d.City, d.Address, d.Region, d.Email,
	).Scan(&uid); err != nil {
		return "", errors.WithMessage(err, "upsert delivery (tx)")
	}
	return uid, nil
}

func (r *DeliveryRepo) Get(ctx context.Context, d models.Delivery) (*models.Delivery, error) {
	const q = `
		SELECT order_uid, name, phone, zip, city, address, region, email
		FROM deliveries
		WHERE order_uid = $1
	`
	var out models.Delivery
	if err := r.db.GetContext(ctx, &out, q, d.OrderUID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.WithMessage(err, "select delivery")
	}
	return &out, nil
}
