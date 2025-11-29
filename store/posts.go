package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Post struct {
	ID                int64       `json:"id"`
	Title             string      `json:"title"`
	Content           string      `json:"content"`
	UserID            int64       `json:"user_id"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	Likes             int64       `json:"likes"`
	Username          string      `json:"username"`
	FileIDs           []uuid.UUID `json:"file_ids"`
	FileExtensions    []string    `json:"file_extensions"`
	OriginalFilenames []string    `json:"original_filenames"`
	Tags              []string    `json:"tags"`
}

type PostStore struct {
	db DBTX
}

func (s *PostStore) Create(ctx context.Context, title, content string, userID int64) (*Post, error) {
	var post Post

	query := `
	INSERT INTO posts (title, content, user_id)
	VALUES($1, $2, $3)
	RETURNING id, title, content, user_id, created_at, updated_at`

	err := s.db.QueryRow(ctx, query, title, content, userID).Scan(
		&post.ID,
		&post.Title,
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

func (s *PostStore) GetByUserID(ctx context.Context, userID int64) ([]*Post, error) {
	var posts []*Post
	query := `
	SELECT p.id, p.title, p.content, p.user_id, p.created_at, p.updated_at, COUNT(pl.post_id), u.name,
	ARRAY_AGG(pf.file_id) FILTER (WHERE pf.file_id IS NOT NULL) as file_ids,
	ARRAY_AGG(pf.file_extension) FILTER (WHERE pf.file_extension IS NOT NULL) as file_extensions,
	ARRAY_AGG(pf.original_filename) FILTER (WHERE pf.original_filename IS NOT NULL) as original_filenames,
	ARRAY_AGG(DISTINCT pt.tag_name) FILTER (WHERE pt.tag_name IS NOT NULL) as tags
	FROM posts p
	LEFT JOIN users u on u.id = p.user_id
	LEFT JOIN post_files pf ON pf.post_id = p.id
	LEFT JOIN post_likes pl ON pl.post_id = p.id
	LEFT JOIN post_tags pt ON pt.post_id = p.id
	WHERE p.user_id = $1
	GROUP BY p.id, u.name`

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	for rows.Next() {
		var post Post
		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.UserID,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Likes,
			&post.Username,
			&post.FileIDs,
			&post.FileExtensions,
			&post.OriginalFilenames,
			&post.Tags,
		); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}

	return posts, nil
}

func (s *PostStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	var post Post
	query := `
	SELECT p.id, p.title, p.content, p.user_id, p.created_at, p.updated_at, COUNT(pl.post_id), u.name,
	ARRAY_AGG(pf.file_id) FILTER (WHERE pf.file_id IS NOT NULL) as file_ids,
	ARRAY_AGG(pf.file_extension) FILTER (WHERE pf.file_extension IS NOT NULL) as file_extensions,
	ARRAY_AGG(pf.original_filename) FILTER (WHERE pf.original_filename IS NOT NULL) as original_filenames,
	ARRAY_AGG(DISTINCT pt.tag_name) FILTER (WHERE pt.tag_name IS NOT NULL) as tags
	FROM posts p
	LEFT JOIN users u on u.id = p.user_id
	LEFT JOIN post_files pf ON pf.post_id = p.id
	LEFT JOIN post_likes pl ON pl.post_id = p.id
	LEFT JOIN post_tags pt ON pt.post_id = p.id
	WHERE p.id = $1
	GROUP BY p.id, u.name`

	err := s.db.QueryRow(ctx, query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.UserID,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Likes,
		&post.Username,
		&post.FileIDs,
		&post.FileExtensions,
		&post.OriginalFilenames,
		&post.Tags,
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
	Post         Post     `json:"post"`
	CommentCount int64    `json:"comment_count"`
	ImageLinks   []string `json:"image_links"`
}

func (s *PostStore) GetUserFeed(ctx context.Context, userID, limit, offset int64) ([]*PostWithMetadata, error) {
	var postsWithMetadata []*PostWithMetadata
	query := `
	SELECT
    	p.id,
     	p.title,
     	p.content,
      	p.user_id,
       	u.name,
       	p.created_at,
        p.updated_at,
        COUNT(DISTINCT pl.post_id) AS like_count,
        COUNT(DISTINCT c.id) AS comment_count,
        ARRAY_AGG(pf.file_id) FILTER (WHERE pf.file_id IS NOT NULL) as file_ids,
		ARRAY_AGG(pf.file_extension) FILTER (WHERE pf.file_extension IS NOT NULL) as file_extensions,
		ARRAY_AGG(pf.original_filename) FILTER (WHERE pf.original_filename IS NOT NULL) as original_filenames,
		ARRAY_AGG(DISTINCT pt.tag_name) FILTER (WHERE pt.tag_name IS NOT NULL) as tags
    FROM posts p
    LEFT JOIN users u ON u.id = p.user_id
    LEFT JOIN comments c ON c.post_id = p.id
    LEFT JOIN post_likes pl ON pl.post_id = p.id
    LEFT JOIN post_tags pt ON pt.post_id = p.id
    LEFT JOIN post_files pf ON pf.post_id = p.id
    WHERE
    	p.user_id = $1
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
			&postWithMetadata.Post.Title,
			&postWithMetadata.Post.Content,
			&postWithMetadata.Post.UserID,
			&postWithMetadata.Post.Username,
			&postWithMetadata.Post.CreatedAt,
			&postWithMetadata.Post.UpdatedAt,
			&postWithMetadata.Post.Likes,
			&postWithMetadata.CommentCount,
			&postWithMetadata.Post.FileIDs,
			&postWithMetadata.Post.FileExtensions,
			&postWithMetadata.Post.OriginalFilenames,
			&postWithMetadata.Post.Tags,
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
		p.title,
     	p.content,
      	p.user_id,
       	u.name,
       	p.created_at,
        p.updated_at,
        COUNT(DISTINCT pl.post_id) AS like_count,
        COUNT(DISTINCT c.id) AS comment_count,
        ARRAY_AGG(pf.file_id) FILTER (WHERE pf.file_id IS NOT NULL) as file_ids,
		ARRAY_AGG(pf.file_extension) FILTER (WHERE pf.file_extension IS NOT NULL) as file_extensions,
		ARRAY_AGG(pf.original_filename) FILTER (WHERE pf.original_filename IS NOT NULL) as original_filenames,
		ARRAY_AGG(DISTINCT pt.tag_name) FILTER (WHERE pt.tag_name IS NOT NULL) as tags
    FROM posts p
    LEFT JOIN users u ON u.id = p.user_id
    LEFT JOIN comments c ON c.post_id = p.id
    LEFT JOIN post_likes pl ON pl.post_id = p.id
    LEFT JOIN post_tags pt ON pt.post_id = p.id
    LEFT JOIN post_files pf ON pf.post_id = p.id
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
			&postWithMetadata.Post.Title,
			&postWithMetadata.Post.Content,
			&postWithMetadata.Post.UserID,
			&postWithMetadata.Post.Username,
			&postWithMetadata.Post.CreatedAt,
			&postWithMetadata.Post.UpdatedAt,
			&postWithMetadata.Post.Likes,
			&postWithMetadata.CommentCount,
			&postWithMetadata.Post.FileIDs,
			&postWithMetadata.Post.FileExtensions,
			&postWithMetadata.Post.OriginalFilenames,
			&postWithMetadata.Post.Tags,
		); err != nil {
			return nil, err
		}
		postsWithMetadata = append(postsWithMetadata, &postWithMetadata)
	}

	return postsWithMetadata, nil
}

func (s *PostStore) GetByTag(ctx context.Context, tagName string, limit, offset int) ([]*Post, error) {
	query := `
	SELECT p.id, p.title, p.content, p.user_id, p.created_at, p.updated_at, COUNT(pl.post_id), u.name,
	ARRAY_AGG(pf.file_id) FILTER (WHERE pf.file_id IS NOT NULL) as file_ids,
	ARRAY_AGG(pf.file_extension) FILTER (WHERE pf.file_extension IS NOT NULL) as file_extensions,
	ARRAY_AGG(pf.original_filename) FILTER (WHERE pf.original_filename IS NOT NULL) as original_filenames,
	ARRAY_AGG(DISTINCT pt.tag_name) FILTER (WHERE pt.tag_name IS NOT NULL) as tags
	FROM posts p
	LEFT JOIN users u on u.id = p.user_id
	LEFT JOIN post_files pf ON pf.post_id = p.id
	LEFT JOIN post_likes pl ON pl.post_id = p.id
	LEFT JOIN post_tags pt ON pt.post_id = p.id
	WHERE pt.tag_name = $1
	GROUP BY p.id, u.name
	LIMIT $2 OFFSET $3`

	rows, err := s.db.Query(ctx, query, tagName, limit, offset)
	if err != nil {
		return nil, err
	}

	var posts []*Post

	for rows.Next() {
		var post Post
		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.UserID,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Likes,
			&post.Username,
			&post.FileIDs,
			&post.FileExtensions,
			&post.OriginalFilenames,
			&post.Tags,
		); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}

	return posts, nil
}
