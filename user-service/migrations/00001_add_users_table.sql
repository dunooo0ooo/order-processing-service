-- +goose Up
-- +goose StatementBegin

CREATE TYPE user_role AS ENUM ('user', 'admin');

CREATE TABLE users
(
    id         BIGSERIAL PRIMARY KEY,
    username   text UNIQUE NOT NULL,
    password   text        NOT NULL,
    role       user_role        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS user_role;
-- +goose StatementEnd
