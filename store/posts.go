package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	UserID    int64     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Likes     int64     `json:"likes"`
	Username  string    `json:"username"`
}

type PostStore struct {
	db DBTX
}

func (s *PostStore) Create(ctx context.Context, content string, userID int64) (*Post, error) {
	var post Post

	query := `
	INSERT INTO posts (content, user_id)
	VALUES($1, $2)
	RETURNING id, content, user_id, created_at, updated_at`

	err := s.db.QueryRow(ctx, query, content, userID).Scan(
		&post.ID,
		&post.Content,
		&post.UserID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (s *PostStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	var post Post
	query := `
	SELECT p.id, p.content, p.user_id, p.created_at, p.updated_at, COUNT(pl.post_id), u.name
	FROM posts p
	LEFT JOIN users u on u.id = p.user_id
	LEFT JOIN post_likes pl ON pl.post_id = p.id
	WHERE p.id = $1
	GROUP BY p.id`

	err := s.db.QueryRow(ctx, query, id).Scan(
		&post.ID,
		&post.Content,
		&post.UserID,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Likes,
		&post.Username,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (s *PostStore) Update(ctx context.Context, content string, id int64) (*Post, error) {
	var post Post
	query := `
	UPDATE posts
	SET content = $1
	WHERE id = $2
	RETURNING id, content, created_at, updated_at`

	err := s.db.QueryRow(ctx, query, content, id).Scan(
		&post.ID,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (s *PostStore) Delete(ctx context.Context, id int64) error {
	query := `
	DELETE FROM posts
	WHERE id = $1`

	result, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

type PostWithMetadata struct {
	Post         Post  `json:"post"`
	CommentCount int64 `json:"comment_count"`
}

func (s *PostStore) GetUserFeed(ctx context.Context, userID, limit, offset int64) ([]*PostWithMetadata, error) {
	var postsWithMetadata []*PostWithMetadata
	query := `
	SELECT
    	p.id,
     	p.content,
      	p.user_id,
       	u.name,
       	p.created_at,
        p.updated_at,
        COUNT(DISTINCT pl.post_id) AS like_count,
        COUNT(DISTINCT c.id) AS comment_count
    FROM posts p
    LEFT JOIN users u ON u.id = p.user_id
    LEFT JOIN comments c ON c.post_id = p.id
    LEFT JOIN post_likes pl ON pl.post_id = p.id
    WHERE
    	p.user_id = $1
     	OR p.user_id IN (
        	SELECT user_id
         	FROM followers
          	WHERE follower_id = $1
        )
    GROUP BY p.id, p.content, p.user_id, u.name, p.created_at, p.updated_at
    ORDER BY p.created_at DESC
    LIMIT $2 OFFSET $3;`

	rows, err := s.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var postWithMetadata PostWithMetadata
		if err := rows.Scan(
			&postWithMetadata.Post.ID,
			&postWithMetadata.Post.Content,
			&postWithMetadata.Post.UserID,
			&postWithMetadata.Post.Username,
			&postWithMetadata.Post.CreatedAt,
			&postWithMetadata.Post.UpdatedAt,
			&postWithMetadata.Post.Likes,
			&postWithMetadata.CommentCount,
		); err != nil {
			return nil, err
		}
		postsWithMetadata = append(postsWithMetadata, &postWithMetadata)
	}

	return postsWithMetadata, nil
}

func (s *PostStore) GetPublicFeed(ctx context.Context, limit, offset int64) ([]*PostWithMetadata, error) {
	var postsWithMetadata []*PostWithMetadata
	query := `
	SELECT
		p.id,
     	p.content,
      	p.user_id,
       	u.name,
       	p.created_at,
        p.updated_at,
        COUNT(DISTINCT pl.post_id) AS like_count,
        COUNT(DISTINCT c.id) AS comment_count
    FROM posts p
    LEFT JOIN users u ON u.id = p.user_id
    LEFT JOIN comments c ON c.post_id = p.id
    LEFT JOIN post_likes pl ON pl.post_id = p.id
    GROUP BY p.id, p.content, p.user_id, u.name, p.created_at, p.updated_at
    ORDER BY like_count DESC, p.created_at DESC
    LIMIT $1 OFFSET $2`

	rows, err := s.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var postWithMetadata PostWithMetadata
		if err := rows.Scan(
			&postWithMetadata.Post.ID,
			&postWithMetadata.Post.Content,
			&postWithMetadata.Post.UserID,
			&postWithMetadata.Post.Username,
			&postWithMetadata.Post.CreatedAt,
			&postWithMetadata.Post.UpdatedAt,
			&postWithMetadata.Post.Likes,
			&postWithMetadata.CommentCount,
		); err != nil {
			return nil, err
		}
		postsWithMetadata = append(postsWithMetadata, &postWithMetadata)
	}

	return postsWithMetadata, nil
}
