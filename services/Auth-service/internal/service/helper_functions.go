package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/TheAmgadX/moltaqa-backend/services/Auth-service/internal/domain"
	userspb "github.com/TheAmgadX/moltaqa-backend/shared/proto/users"
)

const (
	otpDigits         = 6
	maxOTPAttempts    = 3
	refreshTokenBytes = 32
	defaultRefreshTTL = 30 * 24 * time.Hour
)

func (s *AuthService) startOTPTransaction(ctx context.Context, otpTx domain.OTPTransaction, action domain.AuthAction) error {
	otpTx.Action = action

	if err := validateRecipient(otpTx); err != nil {
		return err
	}

	if err := validateAuthAction(otpTx.Action); err != nil {
		return err
	}

	otp, err := generateOTP()
	if err != nil {
		return err
	}

	otpTx.OTPHash = hashValue(otp)
	otpTx.Attempts = 0

	if err := s.otpStore.Save(ctx, otpTx); err != nil {
		return err
	}

	return s.publishOTP(ctx, otpTx, otp)
}

func (s *AuthService) publishOTP(ctx context.Context, otpTx domain.OTPTransaction, otp string) error {
	if s.publisher == nil {
		return nil
	}

	event := domain.NewEvent(otpTx.Email, otpTx.Phone, otp, otpTx.Action)
	if otpTx.Email != "" {
		return s.publisher.PublishSendEmail(ctx, event)
	}

	return s.publisher.PublishSendSMS(ctx, event)
}

func (s *AuthService) recordFailedOTPAttempt(ctx context.Context, otpTx *domain.OTPTransaction) error {
	otpTx.Attempts++
	if err := s.otpStore.Save(ctx, *otpTx); err != nil {
		return err
	}

	if otpTx.Attempts >= maxOTPAttempts {
		return domain.ErrOTPMaxAttemptsExceeded
	}

	return nil
}

func (s *AuthService) createTokenPair(ctx context.Context, userID string) (string, string, error) {
	if s.signer == nil {
		return "", "", domain.ErrServiceUnavailable
	}

	accessToken, err := s.signer.Sign(userID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return "", "", err
	}

	now := time.Now()
	token := domain.NewRefreshToken(userID, hashValue(refreshToken), now, now.Add(defaultRefreshTTL))
	if err := s.repo.CreateRefreshToken(ctx, &token); err != nil {
		return "", "", domain.MapPostgresErrorToDomain(err)
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) resolveUserID(ctx context.Context, otpTx domain.OTPTransaction) (string, error) {
	if s.users == nil {
		return "", domain.ErrServiceUnavailable
	}

	req := &userspb.UserExistsRequest{}
	if otpTx.Email != "" {
		req.Lookup = &userspb.UserExistsRequest_Email{Email: otpTx.Email}
	} else {
		req.Lookup = &userspb.UserExistsRequest_Phone{Phone: otpTx.Phone}
	}

	res, err := s.users.UserExists(ctx, req)
	if err != nil {
		return "", err
	}

	if res.GetResponse() == nil || !res.GetResponse().GetExists() {
		return "", domain.ErrInvalidRecipient
	}

	return res.GetResponse().GetUserId(), nil
}

func sameRecipient(stored domain.OTPTransaction, input domain.OTPTransaction) bool {
	return stored.Email == input.Email && stored.Phone == input.Phone
}

func generateOTP() (string, error) {
	max := big.NewInt(1_000_000)
	value, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%0*d", otpDigits, value.Int64()), nil
}

func generateRefreshToken() (string, error) {
	token := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(token), nil
}

func hashValue(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}
