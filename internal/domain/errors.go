package domain

import "errors"

var (
	ErrInvalidEmail      = errors.New("invalid email address")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidUserID     = errors.New("invalid user ID")

	ErrInvalidFeedName   = errors.New("invalid feed name")
	ErrInvalidFeedURL    = errors.New("invalid feed URL")
	ErrInvalidFeedID     = errors.New("invalid feed ID")
	ErrFeedNotFound      = errors.New("feed not found")
	ErrFeedAlreadyExists = errors.New("feed already exists for this user")
	ErrUnauthorizedFeed  = errors.New("unauthorized to access this feed")

	ErrInvalidFeedItemTitle = errors.New("invalid feed item title")
	ErrInvalidFeedItemLink  = errors.New("invalid feed item link")
	ErrFeedItemNotFound     = errors.New("feed item not found")

	ErrInvalidOTP       = errors.New("invalid OTP")
	ErrInvalidOTPExpiry = errors.New("invalid OTP expiry time")
	ErrOTPExpired       = errors.New("OTP has expired")
	ErrOTPNotFound      = errors.New("OTP not found")

	ErrDatabaseConnection = errors.New("database connection error")
	ErrDatabaseQuery      = errors.New("database query error")
	ErrDuplicateEntry     = errors.New("duplicate entry")
)