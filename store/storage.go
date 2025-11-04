package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource already exists")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Posts interface {
		Create(ctx context.Context, title, content string) (*Post, error)
	}
	Users interface {
		Create(ctx context.Context, user *User) error
		GetByEmail(ctx context.Context, email string) (*User, error)
		GetByName(ctx context.Context, name string) (*User, error)
		GetByID(ctx context.Context, id int64) (*User, error)
	}
}

func NewStorage(db *pgxpool.Pool) Storage {
	return Storage{
		Posts: &PostStore{db},
		Users: &UserStore{db},
	}
}
