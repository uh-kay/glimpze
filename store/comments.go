package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CommentStore struct {
	db DBTX
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

func (s *CommentStore) Create(ctx context.Context, content string, userID, postID int64) (*Comment, error) {
	var comment Comment
	query := `
	INSERT INTO comments (content, user_id, post_id)
	VALUES ($1, $2, $3)
	RETURNING id, post_id, user_id, parent_comment_id, content, likes, created_at, updated_at`

	err := s.db.QueryRow(ctx, query, content, userID, postID).Scan(
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

func (s *CommentStore) Update(ctx context.Context, content string, commentID int64) (*Comment, error) {
	var comment Comment
	query := `
	UPDATE comments
	SET content = $1
	WHERE id = $2
	RETURNING id, post_id, user_id, parent_comment_id, content, likes, created_at, updated_at`

	err := s.db.QueryRow(ctx, query, content, commentID).Scan(
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

func (s *CommentStore) Delete(ctx context.Context, commentID int64) error {
	query := `
	DELETE FROM comments
	WHERE id = $1`

	result, err := s.db.Exec(ctx, query, commentID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *CommentStore) List(ctx context.Context, postID int64, sortBy string, limit, offset int64) (int64, []*Comment, error) {
	var count int64
	query := `SELECT COUNT(*) FROM comments WHERE post_id = $1`
	if err := s.db.QueryRow(ctx, query, postID).Scan(&count); err != nil {
		return -1, nil, err
	}

	var comments []*Comment
	query = `
	SELECT id, post_id, user_id, parent_comment_id, content, likes, created_at, updated_at
	FROM comments
	WHERE post_id = $1`

	switch sortBy {
	case "oldest":
		query += " ORDER BY created_at ASC"
	case "popular":
		query += " ORDER BY likes DESC"
	default:
		query += " ORDER BY created_at DESC"
	}

	query += " LIMIT $2 OFFSET $3"

	rows, err := s.db.Query(ctx, query, postID, limit, offset)
	if err != nil {
		return -1, nil, err
	}

	for rows.Next() {
		var comment Comment
		if err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.ParentCommentID,
			&comment.Content,
			&comment.Likes,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		); err != nil {
			return -1, nil, err
		}

		comments = append(comments, &comment)
	}

	return count, comments, nil
}
