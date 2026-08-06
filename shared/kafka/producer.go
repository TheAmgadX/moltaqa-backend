package kafka

import (
	"context"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

type Producer struct {
	client *kgo.Client
}

func NewProducer(client *kgo.Client) *Producer {
	return &Producer{
		client: client,
	}
}

func (p *Producer) Produce(ctx context.Context, topic Topic, key string, event proto.Message) error {
	payload, err := proto.Marshal(event)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	record := &kgo.Record{
		Topic: string(topic),
		Key:   []byte(key),
		Value: payload,
	}

	return p.client.ProduceSync(ctx, record).FirstErr()
}
