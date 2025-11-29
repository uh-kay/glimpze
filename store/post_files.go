package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PostFileStore struct {
	db DBTX
}

type PostFile struct {
	FileID           uuid.UUID `json:"file_id"`
	FileExtension    string    `json:"file_extension"`
	OriginalFilename string    `json:"original_filename"`
	PostID           int64     `json:"post_id"`
	CreatedAt        time.Time `json:"created_at"`
}

func (s *PostFileStore) Create(ctx context.Context, fileID uuid.UUID, fileExtension, originalFilename string, postID int64) (*PostFile, error) {
	var postFile PostFile
	query := `
	INSERT INTO post_files(file_id, file_extension, original_filename, post_id)
	VALUES($1, $2, $3, $4)
	RETURNING file_id, file_extension, original_filename, post_id, created_at`

	err := s.db.QueryRow(ctx, query, fileID, fileExtension, originalFilename, postID).Scan(
		&postFile.FileID,
		&postFile.FileExtension,
		&postFile.OriginalFilename,
		&postFile.PostID,
		&postFile.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &postFile, nil
}

func (s *PostFileStore) GetByPostID(ctx context.Context, postID int64) ([]*PostFile, error) {

	var postFiles []*PostFile

	query := `
	SELECT file_id, file_extension, original_filename, post_id, created_at
	FROM post_files
	WHERE post_id = $1`

	rows, err := s.db.Query(ctx, query, postID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var postFile PostFile
		if err := rows.Scan(
			&postFile.FileID,
			&postFile.FileExtension,
			&postFile.OriginalFilename,
			&postFile.PostID,
			&postFile.CreatedAt,
		); err != nil {
			return nil, err
		}
		postFiles = append(postFiles, &postFile)
	}

	return postFiles, nil
}

// func (s *PostFileStore) Update(ctx context.Context, tx pgx.Tx, fileID uuid.UUID, fileExtension, originalFilename string, postID int64) (*PostFile, error) {
// 	var postFile PostFile
// 	query := `
// 	UPDATE post_files
// 	SET file_id = $1, file_extension = $2, original_filename = $3
// 	WHERE post_id = $4
// 	RETURNING file_id, file_extension, original_filename, post_id, created_at`

// 	if err := tx.QueryRow(ctx, query, fileID, fileExtension, originalFilename, postID).Scan(
// 		&postFile.FileID,
// 		&postFile.FileExtension,
// 		&postFile.OriginalFilename,
// 		&postFile.PostID,
// 		&postFile.CreatedAt,
// 	); err != nil {
// 		return nil, err
// 	}

// 	return &postFile, nil
// }

func (s *PostFileStore) Delete(ctx context.Context, fileID uuid.UUID) error {
	query := `
	DELETE FROM post_files
	WHERE file_id = $1`

	result, err := s.db.Exec(ctx, query, fileID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
