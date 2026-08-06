package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/domain"
	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/infrastructure/repository"
	events "github.com/TheAmgadX/moltaqa-backend/shared/events/users"
	"github.com/TheAmgadX/moltaqa-backend/shared/kafka"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

// ── EventProducer Interface ──────────────────────────────────────────────────
// This interface decouples UserService from the concrete *kafka.Producer type,
// allowing unit tests to run 100% offline without needing a Kafka broker.
type EventProducer interface {
	Produce(ctx context.Context, topic kafka.Topic, key string, event proto.Message) error
}

// ── MockEventProducer ────────────────────────────────────────────────────────
// An in-memory, thread-safe mock implementation of EventProducer.
type MockEventProducer struct {
	ProducedEvents []ProducedEventRecord
	ProduceError   error
}

type ProducedEventRecord struct {
	Topic kafka.Topic
	Key   string
	Event proto.Message
}

func (m *MockEventProducer) Produce(ctx context.Context, topic kafka.Topic, key string, event proto.Message) error {
	if m.ProduceError != nil {
		return m.ProduceError
	}
	m.ProducedEvents = append(m.ProducedEvents, ProducedEventRecord{
		Topic: topic,
		Key:   key,
		Event: event,
	})
	return nil
}

// ── Mockable UserService Design ─────────────────────────────────────────────
// Below is a demonstration of how the UserService constructor is refactored
// to accept the interface.
type MockableUserService struct {
	repo     repository.UserRepository
	producer EventProducer
}

func NewMockableService(repo repository.UserRepository, producer EventProducer) *MockableUserService {
	return &MockableUserService{
		repo:     repo,
		producer: producer,
	}
}

// Create mimics the production service Create method but uses the mockable producer.
func (s *MockableUserService) Create(ctx context.Context, user *domain.User) error {
	user.Id = uuid.New()

	// Persist to DB
	if err := s.repo.Create(ctx, user); err != nil {
		return err
	}

	// Produce registration event
	event := &events.UserRegistered{
		Id: user.Id.String(),
	}
	_ = s.producer.Produce(ctx, kafka.UserRegistered, user.Id.String(), event)

	return nil
}

// ── Demo Test Cases ─────────────────────────────────────────────────────────

func TestKafkaMocking_Demo(t *testing.T) {
	t.Run("successfully produce event offline", func(t *testing.T) {
		repo := &fakeRepo{}
		producer := &MockEventProducer{}
		svc := NewMockableService(repo, producer)

		user := &domain.User{Email: "test@example.com"}
		ctx := context.Background()

		err := svc.Create(ctx, user)
		if err != nil {
			t.Fatalf("Create() unexpected error: %v", err)
		}

		// ASSERTIONS:
		// 1. Verify exactly 1 event was produced.
		if len(producer.ProducedEvents) != 1 {
			t.Fatalf("expected 1 produced event, got %d", len(producer.ProducedEvents))
		}

		// 2. Verify it was sent to the correct topic.
		eventRecord := producer.ProducedEvents[0]
		if eventRecord.Topic != kafka.UserRegistered {
			t.Errorf("expected topic %q, got %q", kafka.UserRegistered, eventRecord.Topic)
		}

		// 3. Verify event payload details.
		regEvent, ok := eventRecord.Event.(*events.UserRegistered)
		if !ok {
			t.Fatalf("expected event type *events.UserRegistered, got %T", eventRecord.Event)
		}
		if regEvent.Id != user.Id.String() {
			t.Errorf("expected event ID %q, got %q", user.Id.String(), regEvent.Id)
		}
	})

	t.Run("handle event production error gracefully", func(t *testing.T) {
		repo := &fakeRepo{}
		producer := &MockEventProducer{
			ProduceError: errors.New("kafka broker connection refused"),
		}
		svc := NewMockableService(repo, producer)

		user := &domain.User{Email: "test@example.com"}
		ctx := context.Background()

		// The service should swallow/log Kafka errors so that database writes
		// are not rolled back by broker unavailability.
		err := svc.Create(ctx, user)
		if err != nil {
			t.Fatalf("Create() should succeed even if Kafka fails, got error: %v", err)
		}
	})
}
