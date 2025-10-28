package security

import (
	"crypto/rand"
	"fmt"
)

type OTPGenerator struct{}

func NewOTPGenerator() *OTPGenerator {
	return &OTPGenerator{}
}

func (g *OTPGenerator) Generate() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const otpLength = 10
	
	b := make([]byte, otpLength)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	
	otp := make([]byte, otpLength)
	for i := 0; i < otpLength; i++ {
		otp[i] = charset[int(b[i])%len(charset)]
	}
	
	return string(otp), nil
}