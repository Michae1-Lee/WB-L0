-- +goose Up
CREATE TABLE IF NOT EXISTS orders
(
    order_uid          TEXT PRIMARY KEY,
    track_number       TEXT      NOT NULL,
    entry              TEXT      NOT NULL,
    locale             TEXT      NOT NULL,
    internal_signature TEXT,
    customer_id        TEXT      NOT NULL,
    delivery_service   TEXT      NOT NULL,
    shardkey           TEXT      NOT NULL,
    sm_id              INT       NOT NULL,
    date_created       TIMESTAMP NOT NULL,
    oof_shard          TEXT      NOT NULL
);

CREATE TABLE IF NOT EXISTS deliveries
(
    order_uid TEXT PRIMARY KEY REFERENCES orders (order_uid) ON DELETE CASCADE,
    name      TEXT NOT NULL,
    phone     TEXT NOT NULL,
    zip       TEXT NOT NULL,
    city      TEXT NOT NULL,
    address   TEXT NOT NULL,
    region    TEXT NOT NULL,
    email     TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS payments
(
    order_uid     TEXT PRIMARY KEY REFERENCES orders (order_uid) ON DELETE CASCADE,
    transaction   TEXT   NOT NULL,
    request_id    TEXT,
    currency      TEXT   NOT NULL,
    provider      TEXT   NOT NULL,
    amount        INT    NOT NULL,
    payment_dt    BIGINT NOT NULL,
    bank          TEXT   NOT NULL,
    delivery_cost INT    NOT NULL,
    goods_total   INT    NOT NULL,
    custom_fee    INT    NOT NULL
);

CREATE TABLE IF NOT EXISTS items
(
    id           SERIAL PRIMARY KEY,
    order_uid    TEXT REFERENCES orders (order_uid) ON DELETE CASCADE,
    chrt_id      INT  NOT NULL,
    track_number TEXT NOT NULL,
    price        INT  NOT NULL,
    rid          TEXT NOT NULL,
    name         TEXT NOT NULL,
    sale         INT  NOT NULL,
    size         TEXT NOT NULL,
    total_price  INT  NOT NULL,
    nm_id        INT  NOT NULL,
    brand        TEXT NOT NULL,
    status       INT  NOT NULL
);

-- +goose Down
DROP INDEX IF EXISTS idx_items_order_uid;
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS deliveries;
DROP TABLE IF EXISTS orders;
