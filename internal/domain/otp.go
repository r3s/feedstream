package domain

import "time"

type OTP struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	OTP       string    `json:"otp"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (o *OTP) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

func (o *OTP) IsValid(otpToVerify string) bool {
	return o.OTP == otpToVerify && !o.IsExpired()
}

func (o *OTP) Validate() error {
	if o.Email == "" {
		return ErrInvalidEmail
	}
	if o.OTP == "" {
		return ErrInvalidOTP
	}
	if o.ExpiresAt.IsZero() {
		return ErrInvalidOTPExpiry
	}
	return nil
}