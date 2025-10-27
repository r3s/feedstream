package utils

import (
	"fmt"
	"log"
	"rss-reader/config"
	"strconv"

	"gopkg.in/gomail.v2"
)

// SendEmail sends an email using the configured SMTP server
func SendEmail(cfg *config.Config, to, subject, body string) error {
	log.Printf("Attempting to send email to: %s", to)
	log.Printf("Email subject: %s", subject)
	log.Printf("SMTP configuration - Host: %s, Port: %s, Username: %s, From: %s",
		cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.EmailFrom)

	// Validate required config
	if cfg.SMTPHost == "" {
		return fmt.Errorf("SMTP host is not configured")
	}
	if cfg.SMTPPort == "" {
		return fmt.Errorf("SMTP port is not configured")
	}
	if cfg.SMTPUsername == "" {
		return fmt.Errorf("SMTP username is not configured")
	}
	if cfg.SMTPPassword == "" {
		return fmt.Errorf("SMTP password is not configured")
	}
	if cfg.EmailFrom == "" {
		return fmt.Errorf("email from address is not configured")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", cfg.EmailFrom)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	port, err := strconv.Atoi(cfg.SMTPPort)
	if err != nil {
		log.Printf("Error parsing SMTP port '%s': %v", cfg.SMTPPort, err)
		return fmt.Errorf("invalid SMTP port: %s", cfg.SMTPPort)
	}

	log.Printf("Creating SMTP dialer for %s:%d", cfg.SMTPHost, port)
	d := gomail.NewDialer(cfg.SMTPHost, port, cfg.SMTPUsername, cfg.SMTPPassword)

	// Send the email
	log.Printf("Attempting to dial and send email...")
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send email: %v", err)
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Printf("Email sent successfully to: %s", to)
	return nil
}
