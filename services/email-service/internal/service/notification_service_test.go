package service

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"

	"github.com/TheAmgadX/moltaqa-backend/services/email-service/internal/domain"
)

type fakeEmailSender struct {
	recipient string
	subject   string
	body      string
	err       error
}

func (s *fakeEmailSender) Send(_ context.Context, recipient, subject, body string) error {
	s.recipient = recipient
	s.subject = subject
	s.body = body
	return s.err
}

func TestSendEmailDeliversOTP(t *testing.T) {
	sender := &fakeEmailSender{}
	service := NewNotificationService(sender, log.New(io.Discard, "", 0))

	err := service.SendEmail(context.Background(), domain.Notification{
		Email:  "person@example.com",
		OTP:    "123456",
		Action: "Login",
	})
	if err != nil {
		t.Fatalf("SendEmail() error = %v", err)
	}
	if sender.recipient != "person@example.com" || sender.subject != "Your Moltaqa verification code" {
		t.Fatalf("unexpected email metadata: recipient=%q subject=%q", sender.recipient, sender.subject)
	}
	if sender.body == "" {
		t.Fatal("expected an email body")
	}
}

func TestSendEmailReturnsProviderError(t *testing.T) {
	sender := &fakeEmailSender{err: errors.New("provider unavailable")}
	service := NewNotificationService(sender, log.New(io.Discard, "", 0))

	err := service.SendEmail(context.Background(), domain.Notification{Email: "person@example.com", OTP: "123456"})
	if !errors.Is(err, sender.err) {
		t.Fatalf("SendEmail() error = %v, want provider error", err)
	}
}

func TestSendSMSDoesNotRequireProvider(t *testing.T) {
	service := NewNotificationService(nil, log.New(io.Discard, "", 0))

	err := service.SendSMS(context.Background(), domain.Notification{Phone: "+201234567890", OTP: "123456"})
	if err != nil {
		t.Fatalf("SendSMS() error = %v", err)
	}
}
