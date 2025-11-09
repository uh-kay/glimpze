package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Likes     int64     `json:"likes"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostStore struct {
	db *pgxpool.Pool
}

func (s *PostStore) Create(ctx context.Context, tx pgx.Tx, content string, userID int64) (*Post, error) {
	var post Post

	query := `
	INSERT INTO posts (content, user_id)
	VALUES($1, $2)
	RETURNING id, content, likes, user_id, created_at, updated_at`

	err := tx.QueryRow(ctx, query, content, userID).Scan(
		&post.ID,
		&post.Content,
		&post.Likes,
		&post.UserID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (s *PostStore) GetByID(ctx context.Context, tx pgx.Tx, id int64) (*Post, error) {
	var post Post
	query := `
	SELECT id, content, likes, user_id, created_at, updated_at
	FROM posts
	WHERE id = $1`

	err := tx.QueryRow(ctx, query, id).Scan(
		&post.ID,
		&post.Content,
		&post.Likes,
		&post.UserID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (s *PostStore) Update(ctx context.Context, tx pgx.Tx, content string, id int64) (*Post, error) {
	var post Post
	query := `
	UPDATE posts
	SET content = $1
	WHERE id = $2
	RETURNING id, content, likes, created_at, updated_at`

	err := tx.QueryRow(ctx, query, content, id).Scan(
		&post.ID,
		&post.Content,
		&post.Likes,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (s *PostStore) Delete(ctx context.Context, id int64) error {
	query := `
	DELETE FROM posts
	WHERE id = $1`

	result, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
