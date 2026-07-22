package kafka

import (
	"strings"

	"github.com/TheAmgadX/moltaqa-backend/shared/env"
)

// Config contains the common configuration required to create a Kafka client.
// It is shared by both producers and consumers.
type Config struct {
	// Seed brokers used to bootstrap the connection to the Kafka cluster.
	// Example:
	// []string{"kafka-kafka-bootstrap:9092"}
	Brokers []string

	// ClientID uniquely identifies this client instance to the Kafka cluster.
	// Examples:
	// "user-service"
	// "auth-service"
	ClientId string

	// Authentication configuration.
	Username  string
	Password  string
	Mechanism SASLMechanism

	// Consumer group ID for the client.
	GroupID string

	// Whether to use TLS when connecting to brokers.
	TLS bool
}

// SASLMechanism represents the supported SASL authentication mechanisms.
type SASLMechanism string

const (
	SASLMechanismPlain       SASLMechanism = "PLAIN"
	SASLMechanismSCRAMSHA256 SASLMechanism = "SCRAM-SHA-256"
	SASLMechanismSCRAMSHA512 SASLMechanism = "SCRAM-SHA-512"
)

func NewConfig(clientID string, groupID string) *Config {
	return &Config{
		Brokers: strings.Split(
			env.GetString("KAFKA_BROKERS", ""),
			",",
		),

		ClientId: clientID,
		GroupID:  groupID,

		Username: env.GetString("KAFKA_USERNAME", ""),
		Password: env.GetString("KAFKA_PASSWORD", ""),

		Mechanism: SASLMechanism(
			env.GetString("KAFKA_MECHANISM", string(SASLMechanismSCRAMSHA512)),
		),

		TLS: env.GetBool("KAFKA_TLS_ENABLED", false),
	}
}
