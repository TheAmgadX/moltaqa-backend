package grpc

import (
	"context"
	"log"

	"github.com/TheAmgadX/moltaqa-backend/services/auth-service/internal/domain"
	pb "github.com/TheAmgadX/moltaqa-backend/shared/proto/auth"
)

type AuthService interface {
	Login(ctx context.Context, otpTx domain.OTPTransaction) error
	VerifyOTP(ctx context.Context, otpTx domain.OTPTransaction) (string, string, error)
	RestoreAccount(ctx context.Context, otpTx domain.OTPTransaction) error
	DeleteAccount(ctx context.Context, otpTx domain.OTPTransaction) error
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
}

type AuthGRPCServer struct {
	pb.UnimplementedAuthServiceServer
	service AuthService
}

func NewAuthGRPCServer(service AuthService) *AuthGRPCServer {
	return &AuthGRPCServer{service: service}
}

func (s *AuthGRPCServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	otpTx, err := mapLoginRequestToDomain(req)
	if err != nil {
		return nil, mapServiceError(err)
	}

	if err := s.service.Login(ctx, otpTx); err != nil {
		return nil, mapServiceError(err)
	}

	return &pb.LoginResponse{}, nil
}

func (s *AuthGRPCServer) VerifyOTP(ctx context.Context, req *pb.VerifyOTPRequest) (*pb.VerifyOTPResponse, error) {
	otpTx, err := mapVerifyOTPRequestToDomain(req)
	if err != nil {
		log.Printf("mapVerifyOTPRequestToDomain: %v", err)
		return nil, mapServiceError(err)
	}

	accessToken, refreshToken, err := s.service.VerifyOTP(ctx, otpTx)
	if err != nil {
		log.Printf("VerifyOTP service error: %T: %v", err, err)
		return nil, mapServiceError(err)
	}

	return &pb.VerifyOTPResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthGRPCServer) RestoreAccount(ctx context.Context, req *pb.RestoreAccountRequest) (*pb.RestoreAccountResponse, error) {
	otpTx, err := mapRestoreAccountRequestToDomain(req)
	if err != nil {
		return nil, mapServiceError(err)
	}

	if err := s.service.RestoreAccount(ctx, otpTx); err != nil {
		return nil, mapServiceError(err)
	}

	return &pb.RestoreAccountResponse{}, nil
}

func (s *AuthGRPCServer) DeleteAccount(ctx context.Context, req *pb.DeleteAccountRequest) (*pb.DeleteAccountResponse, error) {
	otpTx, err := mapDeleteAccountRequestToDomain(req)
	if err != nil {
		return nil, mapServiceError(err)
	}

	if err := s.service.DeleteAccount(ctx, otpTx); err != nil {
		return nil, mapServiceError(err)
	}

	return &pb.DeleteAccountResponse{}, nil
}

func (s *AuthGRPCServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	accessToken, refreshToken, err := s.service.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, mapServiceError(err)
	}

	return &pb.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
