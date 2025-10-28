package domain

import "time"

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (u *User) Validate() error {
	if u.Email == "" {
		return ErrInvalidEmail
	}
	return nil
}