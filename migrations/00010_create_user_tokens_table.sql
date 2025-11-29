-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_tokens (
    hash bytea PRIMARY KEY,
    user_id bigint not null references users(id) on delete cascade,
    expiry timestamptz not null,
    scope text not null
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_tokens;
-- +goose StatementEnd
