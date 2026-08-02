package grpc

import (
	"github.com/TheAmgadX/moltaqa-backend/services/Auth-service/internal/domain"
	pb "github.com/TheAmgadX/moltaqa-backend/shared/proto/auth"
)

func mapLoginRequestToDomain(req *pb.LoginRequest) (domain.OTPTransaction, error) {
	otpTx := domain.OTPTransaction{
		Action: domain.ActionLogin,
	}

	switch recipient := req.Recipient.(type) {
	case *pb.LoginRequest_Phone:
		otpTx.Phone = recipient.Phone
	case *pb.LoginRequest_Email:
		otpTx.Email = recipient.Email
	default:
		return domain.OTPTransaction{}, domain.ErrInvalidRecipient
	}

	return otpTx, nil
}

func mapRestoreAccountRequestToDomain(req *pb.RestoreAccountRequest) (domain.OTPTransaction, error) {
	otpTx := domain.OTPTransaction{
		Action: domain.ActionRestore,
	}

	switch recipient := req.Recipient.(type) {
	case *pb.RestoreAccountRequest_Phone:
		otpTx.Phone = recipient.Phone
	case *pb.RestoreAccountRequest_Email:
		otpTx.Email = recipient.Email
	default:
		return domain.OTPTransaction{}, domain.ErrInvalidRecipient
	}

	return otpTx, nil
}

func mapDeleteAccountRequestToDomain(req *pb.DeleteAccountRequest) (domain.OTPTransaction, error) {
	otpTx := domain.OTPTransaction{
		Action: domain.ActionDelete,
	}

	switch recipient := req.Recipient.(type) {
	case *pb.DeleteAccountRequest_Phone:
		otpTx.Phone = recipient.Phone
	case *pb.DeleteAccountRequest_Email:
		otpTx.Email = recipient.Email
	default:
		return domain.OTPTransaction{}, domain.ErrInvalidRecipient
	}

	return otpTx, nil
}

func mapVerifyOTPRequestToDomain(req *pb.VerifyOTPRequest) (domain.OTPTransaction, error) {
	otpTx := domain.OTPTransaction{
		OTPHash: req.Otp,
	}

	switch recipient := req.Recipient.(type) {
	case *pb.VerifyOTPRequest_Phone:
		otpTx.Phone = recipient.Phone
	case *pb.VerifyOTPRequest_Email:
		otpTx.Email = recipient.Email
	default:
		return domain.OTPTransaction{}, domain.ErrInvalidRecipient
	}

	return otpTx, nil
}
