-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS followers (
    user_id bigint not null references users(id) on delete cascade,
    follower_id bigint not null references users(id) on delete cascade,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    primary key (user_id, follower_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS followers;
-- +goose StatementEnd
