package email

import (
	"fmt"
	"log"

	"github.com/resend/resend-go/v2"
)

type Service interface {
	SendEmail(to, subject, body string) error
}

type ResendService struct {
	apiKey string
	from   string
	client *resend.Client
}

func NewResendService(apiKey, from string) (*ResendService, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("resend API key is required")
	}
	if from == "" {
		return nil, fmt.Errorf("from email address is required")
	}

	client := resend.NewClient(apiKey)

	return &ResendService{
		apiKey: apiKey,
		from:   from,
		client: client,
	}, nil
}

func (s *ResendService) SendEmail(to, subject, body string) error {
	log.Printf("Attempting to send email to: %s", to)
	log.Printf("Email subject: %s", subject)

	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Html:    body,
		Subject: subject,
	}

	log.Printf("Sending email via Resend API...")

	sent, err := s.client.Emails.Send(params)
	if err != nil {
		log.Printf("Failed to send email via Resend: %v", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email sent successfully to: %s, Message ID: %s", to, sent.Id)
	return nil
}