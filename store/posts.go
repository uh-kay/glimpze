package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Like      int       `json:"like"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PostStore struct {
	db *pgxpool.Pool
}

func (s *PostStore) Create(ctx context.Context, title, content string) (*Post, error) {
	var post Post

	query := `
	INSERT INTO posts (title, content)
	VALUES($1, $2)
	RETURNING id, title, content, created_at, updated_at`

	err := s.db.QueryRow(ctx, query, title, content).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.Like,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &post, nil
}
