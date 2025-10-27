package utils

import (
	"rss-reader/config"
	"strconv"

	"gopkg.in/gomail.v2"
)

// SendEmail sends an email using the configured SMTP server
func SendEmail(cfg *config.Config, to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", cfg.EmailFrom)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	port, _ := strconv.Atoi(cfg.SMTPPort)
	d := gomail.NewDialer(cfg.SMTPHost, port, cfg.SMTPUsername, cfg.SMTPPassword)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
