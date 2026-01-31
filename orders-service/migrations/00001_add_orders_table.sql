-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders
(
    order_uid          text PRIMARY KEY,
    track_number       text,
    entry              text,
    locale             text,
    internal_signature text,
    customer_id        text,
    delivery_service   text,
    shardkey           text,
    sm_id              integer,
    date_created       timestamptz,
    oof_shard          text,
    updated_at         timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_orders_updated_at ON orders (updated_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_orders_updated_at;
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
