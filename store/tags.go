package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TagStore struct {
	db *pgxpool.Pool
}

type Tag struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *TagStore) Create(ctx context.Context, tx pgx.Tx, name string) (*Tag, error) {
	var tag Tag
	query := `
	INSERT INTO tags(name)
	VALUES ($1)
	RETURNING id, name, created_at`

	if err := tx.QueryRow(ctx, query, name).Scan(
		&tag.ID,
		&tag.Name,
		&tag.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &tag, nil
}

func (s *TagStore) GetByID(ctx context.Context, id int64) (*Tag, error) {
	var tag Tag
	query := `
	SELECT id, name, created_at
	FROM tags
	WHERE id = $1`

	if err := s.db.QueryRow(ctx, query, id).Scan(
		&tag.ID,
		&tag.Name,
		&tag.CreatedAt,
	); err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &tag, nil
}

func (s *TagStore) GetByName(ctx context.Context, name string) (*Tag, error) {
	var tag Tag
	query := `
	SELECT id, name, created_at
	FROM tags
	WHERE name = $1`

	if err := s.db.QueryRow(ctx, query, name).Scan(
		&tag.ID,
		&tag.Name,
		&tag.CreatedAt,
	); err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &tag, nil
}

func (s *TagStore) Delete(ctx context.Context, tx pgx.Tx, id int64) error {
	query := `
	DELETE FROM tags
	WHERE id = $1`

	result, err := tx.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
