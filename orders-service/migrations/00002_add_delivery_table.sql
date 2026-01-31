-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS delivery
(
    order_uid text PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    name      text,
    phone     text,
    zip       text,
    city      text,
    address   text,
    region    text,
    email     text
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS delivery;
-- +goose StatementEnd
