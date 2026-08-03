package events

import (
	"context"

	"github.com/TheAmgadX/moltaqa-backend/services/Auth-service/internal/domain"
	"github.com/TheAmgadX/moltaqa-backend/shared/kafka"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type producer interface {
	Produce(ctx context.Context, topic kafka.Topic, key string, event proto.Message) error
}

// Publisher emits auth notification events to the message broker.
type Publisher struct {
	producer producer
}

func NewPublisher(producer producer) *Publisher {
	return &Publisher{producer: producer}
}

func (p *Publisher) PublishSendEmail(ctx context.Context, event domain.Event) error {
	payload, err := buildPayload(event)
	if err != nil {
		return err
	}

	return p.publish(ctx, kafka.AuthSendEmail, event.Email, payload)
}

func (p *Publisher) PublishSendSMS(ctx context.Context, event domain.Event) error {
	payload, err := buildPayload(event)
	if err != nil {
		return err
	}

	return p.publish(ctx, kafka.AuthSendSMS, event.Phone, payload)
}

func (p *Publisher) publish(ctx context.Context, topic kafka.Topic, key string, payload *structpb.Struct) error {
	if p.producer == nil {
		return nil
	}

	return p.producer.Produce(ctx, topic, key, payload)
}

func buildPayload(event domain.Event) (*structpb.Struct, error) {
	return structpb.NewStruct(map[string]any{
		"email":  event.Email,
		"phone":  event.Phone,
		"otp":    event.OTP,
		"action": string(event.Action),
	})
}
