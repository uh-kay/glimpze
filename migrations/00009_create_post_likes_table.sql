-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS post_likes (
    user_id bigint not null references users(id) on delete cascade,
    post_id bigint not null references posts(id) on delete cascade,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    primary key (user_id, post_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS post_likes;
-- +goose StatementEnd
