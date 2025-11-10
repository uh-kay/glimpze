package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostTag struct {
	PostID    int64     `json:"post_id"`
	TagID     int64     `json:"tag_id"`
	TagName   string    `json:"tag_name"`
	CreatedAt time.Time `json:"created_at"`
}

type PostTagStore struct {
	db *pgxpool.Pool
}

func (s *PostTagStore) Create(ctx context.Context, tx pgx.Tx, postID, tagID int64, tagName string) (*PostTag, error) {
	var postTag PostTag
	query := `
	INSERT INTO post_tags (post_id, tag_id, tag_name)
	VALUES ($1, $2, $3)
	RETURNING post_id, tag_id, tag_name, created_at`

	if err := tx.QueryRow(ctx, query, postID, tagID, tagName).Scan(
		&postTag.PostID,
		&postTag.TagID,
		&postTag.TagName,
		&postTag.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &postTag, nil
}

func (s *PostTagStore) Delete(ctx context.Context, tx pgx.Tx, postID, tagID int64) error {
	query := `
	DELETE FROM post_tags
	WHERE post_id = $1 AND tag_id = $2`

	result, err := tx.Exec(ctx, query, postID, tagID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostTagStore) List(ctx context.Context, postID int64) ([]*PostTag, error) {
	var postTags []*PostTag
	query := `
	SELECT post_id, tag_id, tag_name, created_at
	FROM post_tags
	WHERE post_id = $1`

	rows, err := s.db.Query(ctx, query, postID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var postTag PostTag
		if err := rows.Scan(
			&postTag.PostID,
			&postTag.TagID,
			&postTag.TagName,
			&postTag.CreatedAt,
		); err != nil {
			return nil, err
		}
		postTags = append(postTags, &postTag)
	}

	return postTags, nil
}
