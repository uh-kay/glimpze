-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS post_tags (
    post_id bigint not null references posts(id) ON DELETE CASCADE,
    tag_id bigint not null references tags(id) ON DELETE CASCADE,
    tag_name varchar(50) not null references tags(name) ON DELETE CASCADE,
    PRIMARY KEY(post_id, tag_id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS post_tags;
-- +goose StatementEnd
