package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/TheAmgadX/moltaqa-backend/services/Auth-service/internal/domain"
	"github.com/TheAmgadX/moltaqa-backend/services/Auth-service/internal/infrastructure/repository"
	userspb "github.com/TheAmgadX/moltaqa-backend/shared/proto/users"
	"google.golang.org/grpc"
)

type OTPTransactionStore interface {
	Save(ctx context.Context, otpTx domain.OTPTransaction) error
	Get(ctx context.Context, otpTx domain.OTPTransaction) (*domain.OTPTransaction, error)
	Delete(ctx context.Context, otpTx domain.OTPTransaction) error
}

type NotificationPublisher interface {
	PublishSendEmail(ctx context.Context, event domain.Event) error
	PublishSendSMS(ctx context.Context, event domain.Event) error
}

type TokenSigner interface {
	Sign(subject string) (string, error)
}

type UsersClient interface {
	CreateUser(ctx context.Context, in *userspb.CreateUserRequest, opts ...grpc.CallOption) (*userspb.CreateUserResponse, error)
	UserExists(ctx context.Context, in *userspb.UserExistsRequest, opts ...grpc.CallOption) (*userspb.UserExistsResponse, error)
	DeleteUser(ctx context.Context, in *userspb.DeleteUserRequest, opts ...grpc.CallOption) (*userspb.DeleteUserResponse, error)
	RestoreUser(ctx context.Context, in *userspb.RestoreUserRequest, opts ...grpc.CallOption) (*userspb.RestoreUserResponse, error)
}

type AuthService struct {
	repo      repository.AuthRepository
	otpStore  OTPTransactionStore
	publisher NotificationPublisher
	signer    TokenSigner
	users     UsersClient
}

func NewService(
	repo repository.AuthRepository,
	otpStore OTPTransactionStore,
	publisher NotificationPublisher,
	signer TokenSigner,
	users UsersClient,
) (*AuthService, error) {
	return &AuthService{
		repo:      repo,
		otpStore:  otpStore,
		publisher: publisher,
		signer:    signer,
		users:     users,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, otpTx domain.OTPTransaction) error {
	return s.startOTPTransaction(ctx, otpTx, domain.ActionLogin)
}

func (s *AuthService) VerifyOTP(ctx context.Context, otpTx domain.OTPTransaction) (string, string, error) {
	if err := validateOTPRequest(otpTx); err != nil {
		return "", "", err
	}

	log.Println("1. Validation OK")

	hashedOTP := hashValue(otpTx.OTPHash)
	storedOTP, err := s.otpStore.Get(ctx, domain.OTPTransaction{OTPHash: hashedOTP})
	if err != nil {
		log.Printf("otpStore.Get failed: %v", err)

		return "", "", err
	}
	log.Println("2. OTP loaded")

	if storedOTP.Attempts >= maxOTPAttempts {
		return "", "", domain.ErrOTPMaxAttemptsExceeded
	}

	if !sameRecipient(*storedOTP, otpTx) {
		if err := s.recordFailedOTPAttempt(ctx, storedOTP); err != nil {
			return "", "", err
		}

		return "", "", domain.ErrOTPInvalid
	}

	switch storedOTP.Action {
	case domain.ActionLogin:
		userID, err := s.resolveUserID(ctx, *storedOTP)
		if err != nil {
			if errors.Is(err, domain.ErrUserNotFound) {
				log.Printf("user not found: %v", err)

				userID, err = s.createUser(ctx, *storedOTP)
			}
		}
		if err != nil {
			log.Printf("resolve or create user failed: %v", err)

			return "", "", err
		}
		log.Println("3. User resolved")

		accessToken, refreshToken, err := s.createTokenPair(ctx, userID)
		if err != nil {
			log.Printf("createTokenPair failed: %v", err)

			return "", "", err
		}
		log.Println("4. Tokens created")

		if err := s.otpStore.Delete(ctx, *storedOTP); err != nil {
			log.Printf("Delete failed: %v", err)

			return "", "", err
		}

		log.Println("5. OTP deleted")

		return accessToken, refreshToken, nil

	case domain.ActionRestore:
		userID, err := s.resolveUserID(ctx, *storedOTP)
		if err != nil {
			return "", "", err
		}

		if s.users == nil {
			return "", "", domain.ErrServiceUnavailable
		}

		if _, err := s.users.RestoreUser(ctx, &userspb.RestoreUserRequest{Id: userID}); err != nil {
			return "", "", err
		}

		if err := s.otpStore.Delete(ctx, *storedOTP); err != nil {
			log.Printf("Delete failed: %v", err)

			return "", "", err
		}

		return "", "", nil

	case domain.ActionDelete:
		userID, err := s.resolveUserID(ctx, *storedOTP)
		if err != nil {
			return "", "", err
		}

		if s.users == nil {
			return "", "", domain.ErrServiceUnavailable
		}

		if _, err := s.users.DeleteUser(ctx, &userspb.DeleteUserRequest{Id: userID}); err != nil {
			return "", "", err
		}

		if err := s.otpStore.Delete(ctx, *storedOTP); err != nil {
			log.Printf("Delete failed: %v", err)

			return "", "", err
		}

		return "", "", nil

	default:
		return "", "", domain.ErrInvalidAction
	}
}

func (s *AuthService) RestoreAccount(ctx context.Context, otpTx domain.OTPTransaction) error {
	return s.startOTPTransaction(ctx, otpTx, domain.ActionRestore)
}

func (s *AuthService) DeleteAccount(ctx context.Context, otpTx domain.OTPTransaction) error {
	return s.startOTPTransaction(ctx, otpTx, domain.ActionDelete)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	if err := validateRefreshToken(refreshToken); err != nil {
		return "", "", err
	}

	refreshTokenHash := hashValue(refreshToken)
	storedToken, err := s.repo.GetRefreshToken(ctx, refreshTokenHash)
	if err != nil {
		return "", "", domain.MapPostgresErrorToDomain(err)
	}

	if storedToken.RevokedAt != nil {
		return "", "", domain.ErrRefreshTokenInvalid
	}

	if time.Now().After(storedToken.ExpiresAt) {
		return "", "", domain.ErrRefreshTokenExpired
	}

	now := time.Now()
	storedToken.LastUsedAt = now
	storedToken.RevokedAt = &now

	if err := s.repo.UpdateRefreshToken(ctx, storedToken); err != nil {
		return "", "", domain.MapPostgresErrorToDomain(err)
	}

	accessToken, newRefreshToken, err := s.createTokenPair(ctx, storedToken.UserId)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}
