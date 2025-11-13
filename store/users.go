package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	DisplayName string     `json:"display_name"`
	Email       string     `json:"email"`
	Password    password   `json:"-"`
	ActivatedAt *time.Time `json:"activated_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Role        Role       `json:"role"`
	UserLimit   UserLimit  `json:"user_limit"`
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), 12)
	if err != nil {
		return err
	}

	p.text = &text
	p.hash = hash

	return nil
}

func (p *password) Compare(password string) error {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(password))
}

type UserStore struct {
	db *pgxpool.Pool
}

func (s *UserStore) Create(ctx context.Context, tx pgx.Tx, user *User) error {
	query := `
	INSERT INTO users (name, display_name, email, password_hash, role_id, role_name)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, name, display_name, email, activated_at, created_at, updated_at
	`

	err := tx.QueryRow(ctx,
		query,
		user.Name,
		user.DisplayName,
		user.Email,
		user.Password.hash,
		user.Role.ID,
		user.Role.Name,
	).Scan(
		&user.ID,
		&user.Name,
		&user.DisplayName,
		&user.Email,
		&user.ActivatedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User

	query := `
		SELECT users.id, users.name, display_name, email, password_hash, activated_at, users.created_at, users.updated_at,
		       roles.id, roles.name, roles.level, roles.description, roles.created_at,
		       user_limits.user_id, user_limits.comment_limit, user_limits.create_post_limit,
		       user_limits.created_at, user_limits.follow_limit, user_limits.like_limit, user_limits.updated_at
		FROM users
		JOIN roles ON (users.role_id = roles.id)
		JOIN user_limits ON (users.id = user_limits.user_id)
		WHERE email = $1`

	err := s.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.DisplayName,
		&user.Email,
		&user.Password.hash,
		&user.ActivatedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
		&user.Role.CreatedAt,
		&user.UserLimit.UserID,
		&user.UserLimit.CommentLimit,
		&user.UserLimit.CreatePostLimit,
		&user.UserLimit.CreatedAt,
		&user.UserLimit.FollowLimit,
		&user.UserLimit.LikeLimit,
		&user.UserLimit.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) GetByName(ctx context.Context, name string) (*User, error) {
	var user User

	query := `
		SELECT users.id, users.name, display_name, email, password_hash, activated_at, users.created_at, users.updated_at,
		       roles.id, roles.name, roles.level, roles.description, roles.created_at,
		       user_limits.user_id, user_limits.comment_limit, user_limits.create_post_limit,
		       user_limits.created_at, user_limits.follow_limit, user_limits.like_limit, user_limits.updated_at
		FROM users
		JOIN roles ON (users.role_id = roles.id)
		JOIN user_limits ON (users.id = user_limits.user_id)
		WHERE users.name = $1`

	err := s.db.QueryRow(ctx, query, name).Scan(
		&user.ID,
		&user.Name,
		&user.DisplayName,
		&user.Email,
		&user.Password.hash,
		&user.ActivatedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
		&user.Role.CreatedAt,
		&user.UserLimit.UserID,
		&user.UserLimit.CommentLimit,
		&user.UserLimit.CreatePostLimit,
		&user.UserLimit.CreatedAt,
		&user.UserLimit.FollowLimit,
		&user.UserLimit.LikeLimit,
		&user.UserLimit.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) GetByID(ctx context.Context, id int64) (*User, error) {
	var user User

	query := `
		SELECT users.id, users.name, display_name, email, password_hash, activated_at, users.created_at, users.updated_at,
		       roles.id, roles.name, roles.level, roles.description, roles.created_at,
		       user_limits.user_id, user_limits.comment_limit, user_limits.create_post_limit,
		       user_limits.created_at, user_limits.follow_limit, user_limits.like_limit, user_limits.updated_at
		FROM users
		JOIN roles ON (users.role_id = roles.id)
		JOIN user_limits ON (users.id = user_limits.user_id)
		WHERE users.id = $1`

	err := s.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.DisplayName,
		&user.Email,
		&user.Password.hash,
		&user.ActivatedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
		&user.Role.CreatedAt,
		&user.UserLimit.UserID,
		&user.UserLimit.CommentLimit,
		&user.UserLimit.CreatePostLimit,
		&user.UserLimit.CreatedAt,
		&user.UserLimit.FollowLimit,
		&user.UserLimit.LikeLimit,
		&user.UserLimit.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) UpdateRole(ctx context.Context, tx pgx.Tx, name string, role *Role) (*User, error) {
	var user User
	query := `
	WITH updated AS (
		UPDATE users
		SET role_id = $1, role_name = $2
		WHERE name = $3
		RETURNING *
	)
	SELECT
		updated.id,
		updated.name,
		updated.display_name,
		updated.email,
		updated.password_hash,
		updated.activated_at,
		updated.created_at,
		updated.updated_at,
		roles.id as "roles.id",
		roles.name as "roles.name",
		roles.level as "roles.level",
		roles.description as "roles.description"
	FROM updated
	JOIN roles ON (updated.role_id = roles.id)`

	err := tx.QueryRow(ctx, query, role.ID, role.Name, name).Scan(
		&user.ID,
		&user.Name,
		&user.DisplayName,
		&user.Email,
		&user.Password.hash,
		&user.ActivatedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
		&user.Role.Description,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) GetIDs(ctx context.Context, limit, offset int64) ([]int, error) {
	var userIDs []int
	query := `
	SELECT id
	FROM users
	LIMIT $1
	OFFSET $2`

	rows, err := s.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}
