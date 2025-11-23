-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS posts(
    id bigserial PRIMARY KEY,
    title varchar(30) not null,
    content varchar(2048) not null,
    user_id bigint not null references users(id) on delete cascade,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_posts_updated_at
    BEFORE UPDATE ON posts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_posts_updated_at ON posts;
DROP TABLE IF EXISTS posts;
-- +goose StatementEnd
