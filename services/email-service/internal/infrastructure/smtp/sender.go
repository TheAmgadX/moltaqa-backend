package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

type Sender struct {
	config Config
	dialer net.Dialer
}

func NewSender(config Config) (*Sender, error) {
	if config.Host == "" {
		return nil, fmt.Errorf("SMTP_HOST is required")
	}
	if config.Port == "" {
		return nil, fmt.Errorf("SMTP_PORT is required")
	}
	if config.From == "" {
		return nil, fmt.Errorf("SMTP_FROM is required")
	}
	return &Sender{config: config}, nil
}

func (s *Sender) Send(ctx context.Context, recipient, subject, body string) error {
	if strings.ContainsAny(recipient, "\r\n") || strings.ContainsAny(subject, "\r\n") {
		return fmt.Errorf("email headers must not contain newlines")
	}

	connection, err := s.dialer.DialContext(ctx, "tcp", net.JoinHostPort(s.config.Host, s.config.Port))
	if err != nil {
		return fmt.Errorf("dial SMTP server: %w", err)
	}
	defer connection.Close()

	client, err := smtp.NewClient(connection, s.config.Host)
	if err != nil {
		return fmt.Errorf("create SMTP client: %w", err)
	}
	defer client.Quit()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: s.config.Host, MinVersion: tls.VersionTLS12}); err != nil {
			return fmt.Errorf("start SMTP TLS: %w", err)
		}
	}

	if s.config.Username != "" {
		auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("authenticate SMTP client: %w", err)
		}
	}

	if err := client.Mail(s.config.From); err != nil {
		return fmt.Errorf("set email sender: %w", err)
	}
	if err := client.Rcpt(recipient); err != nil {
		return fmt.Errorf("set email recipient: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("open email body: %w", err)
	}
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n", s.config.From, recipient, subject, body)
	if _, err := writer.Write([]byte(message)); err != nil {
		writer.Close()
		return fmt.Errorf("write email body: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("send email body: %w", err)
	}

	return nil
}

func (s *Sender) SetTimeout(timeout time.Duration) {
	s.dialer.Timeout = timeout
}
