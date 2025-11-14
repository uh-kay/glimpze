package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostLikeStore struct {
	db *pgxpool.Pool
}

type PostLike struct {
	UserID    int64     `json:"user_id"`
	PostID    int64     `json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *PostLikeStore) Create(ctx context.Context, tx pgx.Tx, userID, postID int64) (*PostLike, error) {
	var postLike PostLike
	query := `
	INSERT INTO post_likes (user_id, post_id)
	VALUES ($1, $2)
	RETURNING user_id, post_id, created_at`

	err := tx.QueryRow(ctx, query, userID, postID).Scan(
		&postLike.UserID,
		&postLike.PostID,
		&postLike.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &postLike, nil
}

func (s *PostLikeStore) Delete(ctx context.Context, tx pgx.Tx, userID, postID int64) error {
	query := `
	DELETE FROM post_likes
	WHERE user_id = $1 AND post_id = $2`

	result, err := tx.Exec(ctx, query, userID, postID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
