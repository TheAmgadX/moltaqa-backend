package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/TheAmgadX/moltaqa-backend/services/auth-service/internal/domain"
	userspb "github.com/TheAmgadX/moltaqa-backend/shared/proto/users"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	// [TESTING]: Log the OTP so it can be consumed since there are no email/SMS services configured yet.
	fmt.Printf("[TESTING] Generated OTP for %s: %s\n", otpTx.Email+otpTx.Phone, otp)

	otpTx.OTPHash = hashValue(otp)
	otpTx.Attempts = 0

	fmt.Println("[DEBUG] before Save")

	if err := s.otpStore.Save(ctx, otpTx); err != nil {
		return err
	}

	fmt.Println("[DEBUG] before publish")
	err2 := s.publishOTP(ctx, otpTx, otp)
	fmt.Println("[DEBUG] after publish, err:", err2)
	return err2
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
		mappedErr := mapUsersClientError(err)
		logUsersClientError("UserExists", err, mappedErr)
		return "", mappedErr
	}

	if res.GetResponse() == nil || !res.GetResponse().GetExists() {
		return "", domain.ErrUserNotFound
	}

	if res.GetResponse().GetUserId() == "" {
		return "", domain.ErrInternal
	}

	return res.GetResponse().GetUserId(), nil
}

func (s *AuthService) createUser(ctx context.Context, otpTx domain.OTPTransaction) (string, error) {
	if s.users == nil {
		return "", domain.ErrServiceUnavailable
	}

	req := &userspb.CreateUserRequest{}
	if otpTx.Email != "" {
		req.Contact = &userspb.CreateUserRequest_Email{Email: otpTx.Email}
	} else {
		req.Contact = &userspb.CreateUserRequest_Phone{Phone: otpTx.Phone}
	}

	res, err := s.users.CreateUser(ctx, req)
	if err != nil {
		mappedErr := mapUsersClientError(err)
		logUsersClientError("CreateUser", err, mappedErr)
		if errors.Is(mappedErr, domain.ErrAuthAlreadyExists) {
			return s.resolveUserID(ctx, otpTx)
		}

		return "", mappedErr
	}

	if res.GetUser() == nil || res.GetUser().GetId() == "" {
		return "", domain.ErrInternal
	}

	return res.GetUser().GetId(), nil
}

func mapUsersClientError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, domain.ErrUserNotFound) {
		return domain.ErrUserNotFound
	}

	if st, ok := status.FromError(err); ok {
		if isUserNotFoundStatus(st) {
			return domain.ErrUserNotFound
		}

		switch st.Code() {
		case codes.InvalidArgument:
			return domain.ErrInvalidRecipient
		case codes.NotFound:
			return domain.ErrUserNotFound
		case codes.AlreadyExists:
			return domain.ErrAuthAlreadyExists
		case codes.PermissionDenied:
			return domain.ErrPermissionDenied
		case codes.DeadlineExceeded:
			return domain.ErrRequestTimeout
		case codes.Unavailable:
			return domain.ErrServiceUnavailable
		case codes.FailedPrecondition, codes.Aborted:
			return domain.ErrConflict
		default:
			return domain.ErrInternal
		}
	}

	if isUserNotFoundMessage(err.Error()) {
		return domain.ErrUserNotFound
	}

	return domain.ErrInternal
}

func isUserNotFoundStatus(st *status.Status) bool {
	return st.Code() == codes.NotFound || isUserNotFoundMessage(st.Message())
}

func isUserNotFoundMessage(message string) bool {
	return strings.Contains(strings.ToLower(message), domain.ErrUserNotFound.Error())
}

func logUsersClientError(operation string, rawErr error, mappedErr error) {
	if st, ok := status.FromError(rawErr); ok {
		log.Printf("%s failed: users grpc code=%s message=%q mapped=%v", operation, st.Code(), st.Message(), mappedErr)
		return
	}

	log.Printf("%s failed: users raw error=%T: %v mapped=%v", operation, rawErr, rawErr, mappedErr)
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
