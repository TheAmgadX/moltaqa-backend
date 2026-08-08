package events

import (
	"context"
	"fmt"
	"log"

	"github.com/TheAmgadX/moltaqa-backend/services/email-service/internal/domain"
	"github.com/TheAmgadX/moltaqa-backend/services/email-service/internal/service"
	"github.com/TheAmgadX/moltaqa-backend/shared/kafka"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type Consumer struct {
	service *service.NotificationService
	logger  *log.Logger
}

func NewConsumer(service *service.NotificationService, logger *log.Logger) *Consumer {
	return &Consumer{service: service, logger: logger}
}

func (c *Consumer) Handle(ctx context.Context, record *kgo.Record) error {
	notification, err := decodeNotification(record.Value)
	if err != nil {
		// Invalid events cannot succeed on a retry, so acknowledge them after
		// logging instead of preventing later records from being processed.
		if c.logger != nil {
			c.logger.Printf("discarding invalid notification event from %s: %v", record.Topic, err)
		}
		return nil
	}

	switch kafka.Topic(record.Topic) {
	case kafka.AuthSendEmail:
		err = c.service.SendEmail(ctx, notification)
	case kafka.AuthSendSMS:
		err = c.service.SendSMS(ctx, notification)
	default:
		return fmt.Errorf("unsupported notification topic: %s", record.Topic)
	}
	if err != nil {
		return fmt.Errorf("process notification event: %w", err)
	}
	return nil
}

func decodeNotification(payload []byte) (domain.Notification, error) {
	message := &structpb.Struct{}
	if err := proto.Unmarshal(payload, message); err != nil {
		return domain.Notification{}, fmt.Errorf("unmarshal event payload: %w", err)
	}

	fields := message.GetFields()
	return domain.Notification{
		Email:  fields["email"].GetStringValue(),
		Phone:  fields["phone"].GetStringValue(),
		OTP:    fields["otp"].GetStringValue(),
		Action: fields["action"].GetStringValue(),
	}, nil
}
