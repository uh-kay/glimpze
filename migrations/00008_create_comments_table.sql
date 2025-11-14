-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS comments (
    id bigserial primary key,
    post_id bigint not null references posts(id) on delete cascade,
    user_id bigint not null references users(id) on delete cascade,
    parent_comment_id bigint references comments(id) on delete cascade,
    content varchar(2048) not null default '',
    likes bigint not null default 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_comments_updated_at
BEFORE UPDATE ON comments
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_comments_updated_at ON comments;
DROP TABLE IF EXISTS comments;
-- +goose StatementEnd
