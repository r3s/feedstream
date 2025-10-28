package domain

import "time"

type Feed struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (f *Feed) Validate() error {
	if f.Name == "" {
		return ErrInvalidFeedName
	}
	if f.URL == "" {
		return ErrInvalidFeedURL
	}
	if f.UserID <= 0 {
		return ErrInvalidUserID
	}
	return nil
}