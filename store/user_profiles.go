package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type UserProfileStore struct {
	db DBTX
}

type UserProfile struct {
	UserID        int64     `json:"user_id"`
	FileID        uuid.UUID `json:"file_id"`
	FileExtension string    `json:"file_extension"`
	Biodata       string    `json:"biodata"`
	CreatedAt     time.Time `json:"created_at"`
}

func (s *UserProfileStore) Create(ctx context.Context, fileID uuid.UUID, fileExtension string, userID int64, biodata string) (*UserProfile, error) {
	var userProfile UserProfile
	query := `
	INSERT INTO user_profiles (user_id, file_id, file_extension, biodata)
	VALUES ($1, $2, $3, $4)
	RETURNING user_id, file_id, file_extension, biodata, created_at`

	err := s.db.QueryRow(ctx, query, userID, fileID, fileExtension, biodata).Scan(
		&userProfile.UserID,
		&userProfile.FileID,
		&userProfile.FileExtension,
		&userProfile.Biodata,
		&userProfile.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &userProfile, nil
}

func (s *UserProfileStore) GetByUserID(ctx context.Context, userID int64) (*UserProfile, error) {
	var userProfile UserProfile
	query := `
	SELECT user_id, file_id, file_extension, biodata, created_at
	FROM user_profiles
	WHERE user_id = $1`

	err := s.db.QueryRow(ctx, query, userID).Scan(
		&userProfile.UserID,
		&userProfile.FileID,
		&userProfile.FileExtension,
		&userProfile.Biodata,
		&userProfile.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &userProfile, nil
}

func (s *UserProfileStore) Update(ctx context.Context, fileID uuid.UUID, fileExtension string, userID int64, biodata string) (*UserProfile, error) {
	var userProfile UserProfile
	query := `
	UPDATE user_profiles
	SET file_id = $1, file_extension = $2, biodata = $3
	WHERE user_id = $4
	RETURNING user_id, file_id, file_extension, biodata, created_at`

	err := s.db.QueryRow(ctx, query, fileID, fileExtension, biodata, userID).Scan(
		&userProfile.UserID,
		&userProfile.FileID,
		&userProfile.FileExtension,
		&userProfile.Biodata,
		&userProfile.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &userProfile, nil
}
