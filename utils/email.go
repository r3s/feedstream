package utils

import (
	"fmt"
	"log"
	"rss-reader/config"

	"github.com/resend/resend-go/v2"
)

// SendEmail sends an email using the Resend API
func SendEmail(cfg *config.Config, to, subject, body string) error {
	log.Printf("Attempting to send email to: %s", to)
	log.Printf("Email subject: %s", subject)
	log.Printf("Resend configuration - API Key exists: %t, From: %s",
		cfg.ResendAPIKey != "", cfg.EmailFrom)

	// Validate required config
	if cfg.ResendAPIKey == "" {
		return fmt.Errorf("Resend API key is not configured")
	}
	if cfg.EmailFrom == "" {
		return fmt.Errorf("email from address is not configured")
	}

	// Create Resend client
	client := resend.NewClient(cfg.ResendAPIKey)

	// Prepare email parameters
	params := &resend.SendEmailRequest{
		From:    cfg.EmailFrom,
		To:      []string{to},
		Html:    body,
		Subject: subject,
	}

	log.Printf("Sending email via Resend API...")

	// Send the email
	sent, err := client.Emails.Send(params)
	if err != nil {
		log.Printf("Failed to send email via Resend: %v", err)
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Printf("Email sent successfully to: %s, Message ID: %s", to, sent.Id)
	return nil
}
