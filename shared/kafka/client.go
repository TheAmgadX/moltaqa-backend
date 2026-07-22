package kafka

import (
	"crypto/tls"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"
)

func saslMechanism(cfg *Config) (sasl.Mechanism, error) {
	auth := scram.Auth{
		User: cfg.Username,
		Pass: cfg.Password,
	}

	switch cfg.Mechanism {
	case SASLMechanismPlain:
		return plain.Auth{
			User: cfg.Username,
			Pass: cfg.Password,
		}.AsMechanism(), nil

	case SASLMechanismSCRAMSHA256:
		return auth.AsSha256Mechanism(), nil

	case SASLMechanismSCRAMSHA512:
		return auth.AsSha512Mechanism(), nil

	default:
		return nil, fmt.Errorf("unsupported SASL mechanism: %s: ", cfg.Mechanism)
	}
}

func NewClient(cfg *Config) (*kgo.Client, error) {
	// logger := kgo.BasicLogger(os.Stdout, kgo.LogLevelDebug, nil)

	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ClientID(cfg.ClientId),
		// kgo.WithLogger(logger),
	}

	if cfg.GroupID != "" {
		opts = append(opts,
			kgo.ConsumerGroup(cfg.GroupID),
			kgo.AutoCommitMarks(),
		)
	}

	if cfg.Username != "" {
		mech, err := saslMechanism(cfg)
		if err != nil {
			return nil, err
		}
		opts = append(opts, kgo.SASL(mech))
	}

	if cfg.TLS {
		opts = append(opts, kgo.DialTLSConfig(&tls.Config{
			MinVersion: tls.VersionTLS12,
		}))
	}

	return kgo.NewClient(opts...)
}
