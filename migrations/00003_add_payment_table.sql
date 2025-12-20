-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS payment
(
    order_uid     text PRIMARY KEY REFERENCES orders (order_uid) ON DELETE CASCADE,
    transaction   text,
    request_id    text,
    currency      text,
    provider      text,
    amount        integer,
    payment_dt    bigint,
    bank          text,
    delivery_cost integer,
    goods_total   integer,
    custom_fee    integer
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS payment;
-- +goose StatementEnd
