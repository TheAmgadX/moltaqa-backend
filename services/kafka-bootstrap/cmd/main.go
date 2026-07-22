package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/TheAmgadX/moltaqa-backend/shared/kafka"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kerr"
)

const (
	serviceId = "kafka-bootstrap"
)

func createTopics(admin *kadm.Client) {
	ctx := context.Background()

	topics := kafka.Topics()

	results, err := admin.CreateTopics(ctx, 1, 1, nil, topics...)
	if err != nil {
		log.Fatalf("failed to create topics: %v", err)
	}

	for topic, result := range results {
		if result.Err != nil {
			// Ignore "already exists"
			if errors.Is(result.Err, kerr.TopicAlreadyExists) {
				log.Printf("Topic %q already exists", topic)
				continue
			}

			log.Fatalf("failed to create topic %q: %v", topic, result.Err)
		}

		log.Printf("Created topic %q", topic)
	}
}

func main() {
	cfg := kafka.NewConfig(serviceId, "")

	client, err := kafka.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()

	admin := kadm.NewClient(client)
	defer admin.Close()

	fmt.Println("Connected to Kafka!")

	createTopics(admin)

	fmt.Println("Topics are created!")
}
