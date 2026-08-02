package service

import (
	"context"

	"github.com/TheAmgadX/moltaqa-backend/services/Auth-service/internal/domain"
	"github.com/TheAmgadX/moltaqa-backend/services/Auth-service/internal/infrastructure/repository"
)

type AuthService struct {
	repo repository.AuthRepository
}

func NewService(repo repository.AuthRepository) (*AuthService, error) {
	return &AuthService{repo: repo}, nil
}

func (s *AuthService) Login(ctx context.Context, otpTx domain.OTPTransaction) error {
	return nil
}

func (s *AuthService) VerifyOTP(ctx context.Context, otpTx domain.OTPTransaction) (string, string, error) {
	return "", "", nil
}

func (s *AuthService) RestoreAccount(ctx context.Context, otpTx domain.OTPTransaction) error {
	return nil
}

func (s *AuthService) DeleteAccount(ctx context.Context, otpTx domain.OTPTransaction) error {
	return nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	return "", "", nil
}
