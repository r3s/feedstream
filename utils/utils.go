package utils

import (
	"crypto/rand"
	"fmt"
	"log"
)

func GenerateOTP() string {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	// Convert bytes to integer and ensure 6 digits
	num := int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
	if num < 0 {
		num = -num
	}
	// Ensure it's exactly 6 digits (100000-999999)
	otp := 100000 + (num % 900000)
	return fmt.Sprintf("%06d", otp)
}
