package repository

import (
	"database/sql"
	"fmt"
	"rss-reader/internal/domain"
)

type UserRepository interface {
	Create(email string) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	GetByID(id int) (*domain.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(email string) (*domain.User, error) {
	user := &domain.User{Email: email}
	
	err := r.db.QueryRow(
		"INSERT INTO users (email) VALUES ($1) RETURNING id, created_at",
		email,
	).Scan(&user.ID, &user.CreatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	return user, nil
}

func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	user := &domain.User{}
	
	err := r.db.QueryRow(
		"SELECT id, email, created_at FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.CreatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	
	return user, nil
}

func (r *userRepository) GetByID(id int) (*domain.User, error) {
	user := &domain.User{}
	
	err := r.db.QueryRow(
		"SELECT id, email, created_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Email, &user.CreatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	
	return user, nil
}