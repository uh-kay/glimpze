-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL primary key,
    name varchar(255) not null unique,
    level int not null default 0,
    description text,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, description, level)
VALUES ('user', 'can create posts and comments', 1);

INSERT INTO roles (name, description, level)
VALUES ('moderator', 'can update posts and comments', 2);

INSERT INTO roles (name, description, level)
VALUES ('admin', 'can update and delete posts and comments', 3);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS roles;
-- +goose StatementEnd
