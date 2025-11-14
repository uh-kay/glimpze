-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_limits (
    user_id bigint not null references users(id) on delete cascade,
    create_post_limit int not null default 1,
    comment_limit int not null default 3,
    like_limit int not null default 5,
    follow_limit int not null default 50,
    PRIMARY KEY (user_id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_user_limits_updated_at
BEFORE UPDATE ON user_limits
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_user_limits_updated_at ON user_limits;
DROP TABLE IF EXISTS user_limits;
-- +goose StatementEnd
