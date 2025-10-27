package models

import "time"

// User represents a user in the system
type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Feed represents an RSS feed
type Feed struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// FeedItem represents an item in an RSS feed
type FeedItem struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	FeedID      int       `json:"feed_id"`
	FeedName    string    `json:"feed_name"`
	PublishedAt time.Time `json:"published_at"`
	IsNew       bool      `json:"is_new"`
}

// OTP represents a one-time password
type OTP struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	OTP       string    `json:"otp"`
	ExpiresAt time.Time `json:"expires_at"`
}
