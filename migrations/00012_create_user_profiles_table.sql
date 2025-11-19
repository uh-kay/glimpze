-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_profiles (
    user_id bigint not null references users(id) on delete cascade,
    file_id uuid not null,
    file_extension varchar(10) not null,
    biodata varchar(255) not null,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id)
);

CREATE TRIGGER update_user_profiles_updated_at
BEFORE UPDATE ON user_profiles
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_profiles;
-- +goose StatementEnd
