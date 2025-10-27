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
	return fmt.Sprintf("%06d", b)
}
