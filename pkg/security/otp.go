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
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	
	num := int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
	if num < 0 {
		num = -num
	}
	
	otp := 100000 + (num % 900000)
	return fmt.Sprintf("%06d", otp), nil
}