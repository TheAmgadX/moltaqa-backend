package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/TheAmgadX/moltaqa-backend/services/Auth-service/internal/domain"
	goredis "github.com/redis/go-redis/v9"
)

const defaultOTPTransactionTTL = 5 * time.Minute

// OTPTransactionStore persists OTP transactions in Redis with a five-minute expiration.
type OTPTransactionStore struct {
	client *goredis.Client
	prefix string
	ttl    time.Duration
}

func NewOTPTransactionStore(client *goredis.Client, prefix string) *OTPTransactionStore {
	if prefix == "" {
		prefix = "auth:otp"
	}

	return &OTPTransactionStore{
		client: client,
		prefix: prefix,
		ttl:    defaultOTPTransactionTTL,
	}
}

func (s *OTPTransactionStore) Save(ctx context.Context, otpTx domain.OTPTransaction) error {
	payload, err := json.Marshal(otpTx)
	if err != nil {
		return err
	}

	key := s.cacheKey(otpTx)
	if err := s.client.Set(ctx, key, payload, s.ttl).Err(); err != nil {
		return err
	}

	return nil
}

func (s *OTPTransactionStore) Get(ctx context.Context, otpTx domain.OTPTransaction) (*domain.OTPTransaction, error) {
	key := s.cacheKey(otpTx)

	payload, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == goredis.Nil {
			return nil, domain.ErrOTPTransactionNotFound
		}

		return nil, err
	}

	var stored domain.OTPTransaction
	if err := json.Unmarshal(payload, &stored); err != nil {
		return nil, err
	}

	return &stored, nil
}

func (s *OTPTransactionStore) Delete(ctx context.Context, otpTx domain.OTPTransaction) error {
	key := s.cacheKey(otpTx)
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return err
	}

	return nil
}

func (s *OTPTransactionStore) cacheKey(otpTx domain.OTPTransaction) string {
	recipient := otpTx.Email
	recipientType := "email"
	if recipient == "" {
		recipient = otpTx.Phone
		recipientType = "phone"
	}

	if recipient == "" {
		return fmt.Sprintf("%s:%s", s.prefix, otpTx.Action)
	}

	return fmt.Sprintf("%s:%s:%s:%s", s.prefix, otpTx.Action, recipientType, recipient)
}
