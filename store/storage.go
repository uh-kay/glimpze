package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource already exists")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	db    *pgxpool.Pool
	Posts interface {
		Create(ctx context.Context, content string, userID int64) (*Post, error)
		GetByID(ctx context.Context, id int64) (*Post, error)
		Update(ctx context.Context, content string, id int64) (*Post, error)
		Delete(ctx context.Context, id int64) error
	}
	Users interface {
		Create(ctx context.Context, user *User) error
		GetByEmail(ctx context.Context, email string) (*User, error)
		GetByName(ctx context.Context, name string) (*User, error)
		GetByID(ctx context.Context, id int64) (*User, error)
		UpdateRole(ctx context.Context, name string, role *Role) (*User, error)
		GetIDs(ctx context.Context, limit, offset int64) ([]int, error)
	}
	PostFiles interface {
		Create(ctx context.Context, fileID uuid.UUID, fileExtension, originalFilename string, postID int64) (*PostFile, error)
		GetByPostID(ctx context.Context, postID int64) ([]*PostFile, error)
		// Update(ctx context.Context, tx pgx.Tx, fileID uuid.UUID, fileExtension, originalFilename string, postID int64) (*PostFile, error)
		Delete(ctx context.Context, fileID uuid.UUID) error
	}
	Tags interface {
		Create(ctx context.Context, name string) (*Tag, error)
		GetByID(ctx context.Context, id int64) (*Tag, error)
		GetByName(ctx context.Context, name string) (*Tag, error)
		Delete(ctx context.Context, id int64) error
	}
	PostTags interface {
		Create(ctx context.Context, postID, tagID int64, tagName string) (*PostTag, error)
		Delete(ctx context.Context, postID, tagID int64) error
		List(ctx context.Context, postID int64) ([]*PostTag, error)
	}
	Roles interface {
		GetByName(ctx context.Context, name string) (*Role, error)
	}
	Comments interface {
		Create(ctx context.Context, content string, userID, postID int64) (*Comment, error)
		GetByID(ctx context.Context, commentID int64) (*Comment, error)
		Update(ctx context.Context, content string, commentID int64) (*Comment, error)
		Delete(ctx context.Context, commentID int64) error
	}
	UserLimits interface {
		Create(ctx context.Context, userID int64) (*UserLimit, error)
		Add(ctx context.Context, userID int64) (*UserLimit, error)
		Reduce(ctx context.Context, userID int64, limitType string) error
	}
	PostLikes interface {
		Create(ctx context.Context, userID, postID int64) (*PostLike, error)
		Delete(ctx context.Context, userID, postID int64) error
	}
	Followers interface {
		Create(ctx context.Context, userID, followerID int64) (*Follower, error)
	}
}

func NewStorage(db *pgxpool.Pool) Storage {
	return Storage{
		db:         db,
		Posts:      &PostStore{db},
		Users:      &UserStore{db},
		PostFiles:  &PostFileStore{db},
		Tags:       &TagStore{db},
		PostTags:   &PostTagStore{db},
		Roles:      &RoleStore{db},
		Comments:   &CommentStore{db},
		UserLimits: &UserLimitStore{db},
		PostLikes:  &PostLikeStore{db},
		Followers:  &FollowerStore{db},
	}
}

type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

func (s *Storage) WithTx(ctx context.Context, fn func(*Storage) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txStorage := &Storage{
		db:         s.db,
		Users:      &UserStore{db: tx},
		Posts:      &PostStore{db: tx},
		PostFiles:  &PostFileStore{db: tx},
		UserLimits: &UserLimitStore{db: tx},
		Tags:       &TagStore{db: tx},
		PostTags:   &PostTagStore{db: tx},
		Comments:   &CommentStore{db: tx},
		PostLikes:  &PostLikeStore{db: tx},
		Followers:  &FollowerStore{db: tx},
	}

	if err := fn(txStorage); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
