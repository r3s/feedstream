package repository

import (
	"database/sql"
	"fmt"
	"rss-reader/internal/domain"
	"time"
)

type OTPRepository interface {
	Store(email, otp string, expiresAt time.Time) error
	GetLatestByEmail(email string) (*domain.OTP, error)
	DeleteByEmail(email string) error
}

type otpRepository struct {
	db *sql.DB
}

func NewOTPRepository(db *sql.DB) OTPRepository {
	return &otpRepository{db: db}
}

func (r *otpRepository) Store(email, otp string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		"INSERT INTO otps (email, otp, expires_at) VALUES ($1, $2, $3)",
		email, otp, expiresAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}
	
	return nil
}

func (r *otpRepository) GetLatestByEmail(email string) (*domain.OTP, error) {
	otp := &domain.OTP{}
	
	err := r.db.QueryRow(
		"SELECT id, email, otp, expires_at FROM otps WHERE email = $1 ORDER BY expires_at DESC LIMIT 1",
		email,
	).Scan(&otp.ID, &otp.Email, &otp.OTP, &otp.ExpiresAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrOTPNotFound
		}
		return nil, fmt.Errorf("failed to get OTP: %w", err)
	}
	
	return otp, nil
}

func (r *otpRepository) DeleteByEmail(email string) error {
	_, err := r.db.Exec("DELETE FROM otps WHERE email = $1", email)
	
	if err != nil {
		return fmt.Errorf("failed to delete OTP: %w", err)
	}
	
	return nil
}