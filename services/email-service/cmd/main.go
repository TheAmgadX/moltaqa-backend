package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	eventconsumer "github.com/TheAmgadX/moltaqa-backend/services/email-service/internal/infrastructure/events"
	"github.com/TheAmgadX/moltaqa-backend/services/email-service/internal/infrastructure/smtp"
	"github.com/TheAmgadX/moltaqa-backend/services/email-service/internal/service"
	"github.com/TheAmgadX/moltaqa-backend/shared/env"
	"github.com/TheAmgadX/moltaqa-backend/shared/kafka"
)

const (
	serviceID       = "email-service"
	consumerGroupID = "email-service"
)

func main() {
	logger := log.New(os.Stdout, "email-service: ", log.LstdFlags|log.LUTC)

	emailSender, err := smtp.NewSender(smtp.Config{
		Host:     env.GetString("SMTP_HOST", ""),
		Port:     env.GetString("SMTP_PORT", "587"),
		Username: env.GetString("SMTP_USERNAME", ""),
		Password: env.GetString("SMTP_PASSWORD", ""),
		From:     env.GetString("SMTP_FROM", ""),
	})
	if err != nil {
		logger.Fatalf("invalid SMTP configuration: %v", err)
	}
	emailSender.SetTimeout(10 * time.Second)

	kafkaConfig := kafka.NewConfig(serviceID, consumerGroupID)
	kafkaConfig.Topics = []string{kafka.AuthSendEmail.String(), kafka.AuthSendSMS.String()}
	kafkaClient, err := kafka.NewClient(kafkaConfig)
	if err != nil {
		logger.Fatalf("create Kafka client: %v", err)
	}
	defer kafkaClient.Close()

	service := service.NewNotificationService(emailSender, logger)
	handler := eventconsumer.NewConsumer(service, logger)
	errors := make(chan error, 1)
	consumer := kafka.NewConsumer(kafkaClient, errors)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go consumer.Consume(ctx, handler.Handle)

	select {
	case <-ctx.Done():
		logger.Println("shutdown signal received")
	case err := <-errors:
		logger.Printf("Kafka consumer stopped: %v", err)
	}
}
