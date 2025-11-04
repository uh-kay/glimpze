package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	DisplayName string     `json:"display_name"`
	Email       string     `json:"email"`
	Password    password   `json:"-"`
	Role        string     `json:"role"`
	ActivatedAt *time.Time `json:"activated_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
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

func (s *UserStore) Create(ctx context.Context, user *User) error {
	query := `
	INSERT INTO users (id, name, display_name, email, password_hash)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, name, display_name, email, role, activated_at, created_at, updated_at
	`

	err := s.db.QueryRow(ctx,
		query,
		user.ID,
		user.Name,
		user.DisplayName,
		user.Email,
		user.Password.hash,
	).Scan(
		&user.ID,
		&user.Name,
		&user.DisplayName,
		&user.Email,
		&user.Role,
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
	SELECT id, name, display_name, email, password_hash, role, activated_at, created_at, updated_at
	FROM users
	WHERE email = $1`

	err := s.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.DisplayName,
		&user.Email,
		&user.Password.hash,
		&user.Role,
		&user.ActivatedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) GetByName(ctx context.Context, name string) (*User, error) {
	var user User

	query := `
	SELECT id, name, display_name, email, password_hash, role, activated_at, created_at, updated_at
	FROM users
	WHERE name = $1`

	err := s.db.QueryRow(ctx, query, name).Scan(
		&user.ID,
		&user.Name,
		&user.DisplayName,
		&user.Email,
		&user.Password.hash,
		&user.Role,
		&user.ActivatedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) GetByID(ctx context.Context, id int64) (*User, error) {
	var user User

	query := `
	SELECT id, name, display_name, email, password_hash, role, activated_at, created_at, updated_at
	FROM users
	WHERE id = $1`

	err := s.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.DisplayName,
		&user.Email,
		&user.Password.hash,
		&user.Role,
		&user.ActivatedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
