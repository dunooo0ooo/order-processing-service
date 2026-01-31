-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS items
(
    order_uid    text REFERENCES orders (order_uid) ON DELETE CASCADE,
    chrt_id      bigint,
    track_number text,
    price        integer,
    rid          text,
    name         text,
    sale         integer,
    size         text,
    total_price  integer,
    nm_id        bigint,
    brand        text,
    status       integer,
    PRIMARY KEY (order_uid, chrt_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS items;
-- +goose StatementEnd
