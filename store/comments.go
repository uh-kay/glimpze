package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CommentStore struct {
	db *pgxpool.Pool
}

type Comment struct {
	ID              int64          `json:"id"`
	PostID          int64          `json:"post_id"`
	UserID          int64          `json:"user_id"`
	ParentCommentID pgtype.Numeric `json:"parent_comment_id"`
	Content         string         `json:"content"`
	Likes           int64          `json:"likes"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

func (s *CommentStore) Create(ctx context.Context, tx pgx.Tx, content string, userID, postID int64) (*Comment, error) {
	var comment Comment
	query := `
	INSERT INTO comments (content, user_id, post_id)
	VALUES ($1, $2, $3)
	RETURNING id, post_id, user_id, parent_comment_id, content, likes, created_at, updated_at`

	err := tx.QueryRow(ctx, query, content, userID, postID).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.ParentCommentID,
		&comment.Content,
		&comment.Likes,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &comment, nil
}

func (s *CommentStore) GetByID(ctx context.Context, commentID int64) (*Comment, error) {
	var comment Comment
	query := `
	SELECT id, post_id, user_id, parent_comment_id, content, likes, created_at, updated_at
	FROM comments
	WHERE id = $1`

	err := s.db.QueryRow(ctx, query, commentID).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.ParentCommentID,
		&comment.Content,
		&comment.Likes,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &comment, nil
}

func (s *CommentStore) Update(ctx context.Context, tx pgx.Tx, content string, commentID int64) (*Comment, error) {
	var comment Comment
	query := `
	UPDATE comments
	SET content = $1
	WHERE id = $2
	RETURNING id, post_id, user_id, parent_comment_id, content, likes, created_at, updated_at`

	err := tx.QueryRow(ctx, query, content, commentID).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.ParentCommentID,
		&comment.Content,
		&comment.Likes,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &comment, nil
}

func (s *CommentStore) Delete(ctx context.Context, tx pgx.Tx, commentID int64) error {
	query := `
	DELETE FROM comments
	WHERE id = $1`

	result, err := tx.Exec(ctx, query, commentID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
