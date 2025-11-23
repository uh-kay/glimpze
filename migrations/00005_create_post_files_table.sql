-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS post_files(
    file_id uuid PRIMARY KEY not null,
    file_extension varchar(10) not null,
    original_filename varchar(260) not null,
    position int not null,
    post_id bigint not null references posts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS post_files;
-- +goose StatementEnd
