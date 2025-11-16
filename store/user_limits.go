package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type UserLimitStore struct {
	db DBTX
}

type UserLimit struct {
	UserID          int64     `json:"user_id"`
	CreatePostLimit int       `json:"create_post_limit"`
	CommentLimit    int       `json:"comment_limit"`
	LikeLimit       int       `json:"like_limit"`
	FollowLimit     int       `json:"follow_limit"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (s *UserLimitStore) Create(ctx context.Context, userID int64) (*UserLimit, error) {
	var userLimit UserLimit
	query := `
	INSERT INTO user_limits (user_id)
	VALUES ($1)
	RETURNING user_id, create_post_limit, comment_limit, like_limit, follow_limit, created_at, updated_at`

	err := s.db.QueryRow(ctx, query, userID).Scan(
		&userLimit.UserID,
		&userLimit.CreatePostLimit,
		&userLimit.CommentLimit,
		&userLimit.LikeLimit,
		&userLimit.FollowLimit,
		&userLimit.CreatedAt,
		&userLimit.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &userLimit, nil
}

func (s *UserLimitStore) Add(ctx context.Context, userID int64) (*UserLimit, error) {
	var userLimit UserLimit
	query := `
	UPDATE user_limits
	SET
		create_post_limit = create_post_limit + 1,
		comment_limit = comment_limit + 3,
		like_limit = like_limit + 5
	WHERE user_id = $1
	RETURNING user_id, create_post_limit, comment_limit, like_limit, follow_limit, created_at, updated_at
	`

	err := s.db.QueryRow(ctx, query, userID).Scan(
		&userLimit.UserID,
		&userLimit.CreatePostLimit,
		&userLimit.CommentLimit,
		&userLimit.LikeLimit,
		&userLimit.FollowLimit,
		&userLimit.CreatedAt,
		&userLimit.UpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &userLimit, nil
}

func (s *UserLimitStore) Reduce(ctx context.Context, userID int64, limitType string) error {
	validLimits := map[string]bool{
		"create_post_limit": true,
		"comment_limit":     true,
		"like_limit":        true,
		"follow_limit":      true,
	}

	if !validLimits[limitType] {
		return fmt.Errorf("invalid limit type: %s", limitType)
	}

	query := fmt.Sprintf(`
	UPDATE user_limits
	SET %s = %s - 1
	WHERE user_id = $1 AND %s > 0`, limitType, limitType, limitType)

	result, err := s.db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no limit to reduce for user %d", userID)
	}

	return nil
}
