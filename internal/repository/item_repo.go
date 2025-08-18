package repository

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"wb/internal/models"
)

type ItemRepo struct {
	db *sqlx.DB
}

func NewItemRepo(db *sqlx.DB) *ItemRepo {
	return &ItemRepo{
		db: db,
	}
}

func (r *ItemRepo) UpsertTx(ctx context.Context, tx *sqlx.Tx, it models.Item) (int, error) {
	const q = `
		INSERT INTO items (
			order_uid, chrt_id, track_number, price, rid, name, sale, size,
			total_price, nm_id, brand, status
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET
			order_uid=EXCLUDED.order_uid,
			chrt_id=EXCLUDED.chrt_id,
			track_number=EXCLUDED.track_number,
			price=EXCLUDED.price,
			rid=EXCLUDED.rid,
			name=EXCLUDED.name,
			sale=EXCLUDED.sale,
			size=EXCLUDED.size,
			total_price=EXCLUDED.total_price,
			nm_id=EXCLUDED.nm_id,
			brand=EXCLUDED.brand,
			status=EXCLUDED.status
		RETURNING id
	`
	var id int
	if err := tx.QueryRowxContext(ctx, q,
		it.OrderUID, it.ChrtId, it.TrackNumber, it.Price, it.Rid,
		it.Name, it.Sale, it.Size, it.TotalPrice, it.NmId, it.Brand, it.Status,
	).Scan(&id); err != nil {
		return 0, errors.WithMessage(err, "upsert item (tx)")
	}
	return id, nil
}

func (r *ItemRepo) GetByOrderUID(ctx context.Context, orderUID string) ([]models.Item, error) {
	const q = `
		SELECT order_uid, chrt_id, track_number, price, rid, name, sale, size,
		       total_price, nm_id, brand, status
		FROM items
		WHERE order_uid = $1
		ORDER BY id
	`
	var out []models.Item
	if err := r.db.SelectContext(ctx, &out, q, orderUID); err != nil && err != sql.ErrNoRows {
		return nil, errors.WithMessage(err, "select items by order_uid")
	}
	return out, nil
}

func (r *ItemRepo) DeleteByOrderUIDTx(ctx context.Context, tx *sqlx.Tx, orderUID string) (int64, error) {
	const q = `DELETE FROM items WHERE order_uid = $1`
	res, err := tx.ExecContext(ctx, q, orderUID)
	if err != nil {
		return 0, errors.WithMessage(err, "delete items by order_uid")
	}
	affected, _ := res.RowsAffected()
	return affected, nil
}
