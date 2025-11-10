package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource already exists")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Posts interface {
		Create(ctx context.Context, tx pgx.Tx, content string, userID int64) (*Post, error)
		GetByID(ctx context.Context, tx pgx.Tx, id int64) (*Post, error)
		Update(ctx context.Context, tx pgx.Tx, content string, id int64) (*Post, error)
		Delete(ctx context.Context, id int64) error
	}
	Users interface {
		Create(ctx context.Context, user *User) error
		GetByEmail(ctx context.Context, email string) (*User, error)
		GetByName(ctx context.Context, name string) (*User, error)
		GetByID(ctx context.Context, id int64) (*User, error)
	}
	PostFiles interface {
		Create(ctx context.Context, tx pgx.Tx, fileID uuid.UUID, fileExtension, originalFilename string, postID int64) (*PostFile, error)
		GetByPostID(ctx context.Context, tx pgx.Tx, postID int64) ([]*PostFile, error)
		// Update(ctx context.Context, tx pgx.Tx, fileID uuid.UUID, fileExtension, originalFilename string, postID int64) (*PostFile, error)
		Delete(ctx context.Context, tx pgx.Tx, fileID uuid.UUID) error
	}
	Tags interface {
		Create(ctx context.Context, tx pgx.Tx, name string) (*Tag, error)
		GetByID(ctx context.Context, id int64) (*Tag, error)
		GetByName(ctx context.Context, name string) (*Tag, error)
		Delete(ctx context.Context, tx pgx.Tx, id int64) error
	}
	PostTags interface {
		Create(ctx context.Context, tx pgx.Tx, postID, tagID int64, tagName string) (*PostTag, error)
		Delete(ctx context.Context, tx pgx.Tx, postID, tagID int64) error
		List(ctx context.Context, postID int64) ([]*PostTag, error)
	}
}

func NewStorage(db *pgxpool.Pool) Storage {
	return Storage{
		Posts:     &PostStore{db},
		Users:     &UserStore{db},
		PostFiles: &PostFileStore{db},
		Tags:      &TagStore{db},
		PostTags:  &PostTagStore{db},
	}
}
