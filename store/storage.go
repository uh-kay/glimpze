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
		GetByID(ctx context.Context, id int64) (*Post, error)
		Update(ctx context.Context, tx pgx.Tx, content string, id int64) (*Post, error)
		Delete(ctx context.Context, id int64) error
	}
	Users interface {
		Create(ctx context.Context, tx pgx.Tx, user *User) error
		GetByEmail(ctx context.Context, email string) (*User, error)
		GetByName(ctx context.Context, name string) (*User, error)
		GetByID(ctx context.Context, id int64) (*User, error)
		UpdateRole(ctx context.Context, tx pgx.Tx, name string, role *Role) (*User, error)
		GetIDs(ctx context.Context, limit, offset int64) ([]int, error)
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
	Roles interface {
		GetByName(ctx context.Context, name string) (*Role, error)
	}
	Comments interface {
		Create(ctx context.Context, tx pgx.Tx, content string, userID, postID int64) (*Comment, error)
		GetByID(ctx context.Context, commentID int64) (*Comment, error)
		Update(ctx context.Context, tx pgx.Tx, content string, commentID int64) (*Comment, error)
		Delete(ctx context.Context, tx pgx.Tx, commentID int64) error
	}
	UserLimits interface {
		Create(ctx context.Context, tx pgx.Tx, userID int64) (*UserLimit, error)
		Add(ctx context.Context, tx pgx.Tx, userID int64) (*UserLimit, error)
		Reduce(ctx context.Context, tx pgx.Tx, userID int64, limitType string) error
	}
	PostLikes interface {
		Create(ctx context.Context, tx pgx.Tx, userID, postID int64) (*PostLike, error)
		Delete(ctx context.Context, tx pgx.Tx, userID, postID int64) error
	}
}

func NewStorage(db *pgxpool.Pool) Storage {
	return Storage{
		Posts:      &PostStore{db},
		Users:      &UserStore{db},
		PostFiles:  &PostFileStore{db},
		Tags:       &TagStore{db},
		PostTags:   &PostTagStore{db},
		Roles:      &RoleStore{db},
		Comments:   &CommentStore{db},
		UserLimits: &UserLimitStore{db},
		PostLikes:  &PostLikeStore{db},
	}
}
