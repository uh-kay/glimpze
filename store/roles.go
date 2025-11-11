package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleStore struct {
	db *pgxpool.Pool
}

type Role struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Level       int       `json:"level"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *RoleStore) GetByName(ctx context.Context, name string) (*Role, error) {
	var role Role
	query := `
	SELECT id, name, level, description, created_at
	FROM roles
	WHERE name = $1`

	if err := s.db.QueryRow(ctx, query, name).Scan(
		&role.ID,
		&role.Name,
		&role.Level,
		&role.Description,
		&role.CreatedAt,
	); err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &role, nil
}
