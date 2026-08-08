package service

import (
	"context"
	"fmt"
	"log"

	"github.com/TheAmgadX/moltaqa-backend/services/email-service/internal/domain"
)

// EmailSender delivers an email through an SMTP provider.
type EmailSender interface {
	Send(ctx context.Context, recipient, subject, body string) error
}

type NotificationService struct {
	emailSender EmailSender
	logger      *log.Logger
}

func NewNotificationService(emailSender EmailSender, logger *log.Logger) *NotificationService {
	return &NotificationService{emailSender: emailSender, logger: logger}
}

func (s *NotificationService) SendEmail(ctx context.Context, notification domain.Notification) error {
	if err := notification.ValidateEmail(); err != nil {
		return err
	}
	if s.emailSender == nil {
		return fmt.Errorf("email sender is not configured")
	}

	return s.emailSender.Send(
		ctx,
		notification.Email,
		"Your Moltaqa verification code",
		fmt.Sprintf("Your verification code for %s is: %s\n\nThis code expires soon. Do not share it with anyone.", notification.Action, notification.OTP),
	)
}

// SendSMS intentionally has no external provider. It validates and records
// the request so SMS support can be added later without changing the Kafka
// contract or the consumer flow.
func (s *NotificationService) SendSMS(_ context.Context, notification domain.Notification) error {
	if err := notification.ValidateSMS(); err != nil {
		return err
	}
	if s.logger != nil {
		s.logger.Printf("SMS delivery skipped (no provider configured): recipient=%s action=%s", notification.Phone, notification.Action)
	}
	return nil
}
