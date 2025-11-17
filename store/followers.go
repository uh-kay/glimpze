package store

import (
	"context"
	"time"
)

type Follower struct {
	UserID     int64     `json:"user_id"`
	FollowerID int64     `json:"follower_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type FollowerStore struct {
	db DBTX
}

func (s *FollowerStore) Create(ctx context.Context, userID, followerID int64) (*Follower, error) {
	var follower Follower
	query := `
	INSERT INTO followers (user_id, follower_id)
	VALUES ($1, $2)
	RETURNING user_id, follower_id, created_at`

	err := s.db.QueryRow(ctx, query, userID, followerID).Scan(
		&follower.UserID,
		&follower.FollowerID,
		&follower.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &follower, nil
}

func (s *FollowerStore) Delete(ctx context.Context, userID, followerID int64) error {
	query := `
	DELETE FROM followers
	WHERE user_id = $1 AND follower_id = $2`

	result, err := s.db.Exec(ctx, query, userID, followerID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
