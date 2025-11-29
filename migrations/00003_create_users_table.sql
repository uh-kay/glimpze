-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    name varchar(30) unique NOT NULL,
    display_name varchar(30) not null default '',
    email citext UNIQUE NOT NULL,
    role_id int not null references roles(id) on delete cascade,
    role_name varchar(255) not null references roles(name) on delete cascade,
    password_hash bytea NOT NULL,
    activated bool NOT NULL default false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
