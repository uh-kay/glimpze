-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    name varchar(30) unique NOT NULL,
    display_name varchar(30) not null default '',
    email citext UNIQUE NOT NULL,
    role text NOT NULL DEFAULT 'user',
    password_hash bytea NOT NULL,
    activated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
