package grpc

import (
	"context"

	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/service"
	pb "github.com/TheAmgadX/moltaqa-backend/shared/proto/users"
)

type UserGRPCServer struct {
	pb.UnimplementedUsersServiceServer
	service *service.UserService
}

func NewUserGRPCServer(service *service.UserService) *UserGRPCServer {
	return &UserGRPCServer{
		service: service,
	}
}

func (s *UserGRPCServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	user := mapCreateUserRequestToDomain(req)

	err := s.service.Create(ctx, user)

	if err != nil {
		return nil, mapServiceError(err)
	}

	return &pb.CreateUserResponse{
		User: mapUserToProto(user),
	}, nil
}

func (s *UserGRPCServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	user, err := mapUpdateUserRequestToDomain(req)

	if err != nil {
		return nil, mapServiceError(err)
	}

	err = s.service.Update(ctx, user)

	if err != nil {
		return nil, mapServiceError(err)
	}

	return &pb.UpdateUserResponse{}, nil
}

func (s *UserGRPCServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	err := s.service.Delete(ctx, req.Id)

	if err != nil {
		return nil, mapServiceError(err)
	}

	return &pb.DeleteUserResponse{}, nil
}

func (s *UserGRPCServer) RestoreUser(ctx context.Context, req *pb.RestoreUserRequest) (*pb.RestoreUserResponse, error) {
	err := s.service.Restore(ctx, req.Id)

	if err != nil {
		return nil, mapServiceError(err)
	}

	return &pb.RestoreUserResponse{}, nil
}

func (s *UserGRPCServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	lookup, err := mapGetUserRequestLookupToDomain(req)

	if err != nil {
		return nil, mapServiceError(err)
	}

	user, err := s.service.Get(ctx, lookup)

	if err != nil {
		return nil, mapServiceError(err)
	}

	return &pb.GetUserResponse{
		User: mapUserToProto(user),
	}, nil
}

func (s *UserGRPCServer) GetUsers(ctx context.Context, req *pb.GetUsersRequest) (*pb.GetUsersResponse, error) {
	users, err := s.service.GetUsers(ctx, req.UserIds)

	if err != nil {
		return nil, mapServiceError(err)
	}

	return &pb.GetUsersResponse{
		Users: mapUsersToProto(users),
	}, nil
}

func (s *UserGRPCServer) GetUserSummary(ctx context.Context, req *pb.GetUserSummaryRequest) (*pb.GetUserSummaryResponse, error) {
	userSummary, err := s.service.GetUserSummary(ctx, req.Id)

	if err != nil {
		return nil, mapServiceError(err)
	}

	return &pb.GetUserSummaryResponse{
		User: mapUserSummaryToProto(userSummary),
	}, nil
}

func (s *UserGRPCServer) GetUsersSummary(ctx context.Context, req *pb.GetUsersSummaryRequest) (*pb.GetUsersSummaryResponse, error) {
	usersSummary, err := s.service.GetUsersSummary(ctx, req.UserIds)

	if err != nil {
		return nil, mapServiceError(err)
	}

	return &pb.GetUsersSummaryResponse{
		Users: mapUsersSummaryToProto(usersSummary),
	}, nil
}

func (s *UserGRPCServer) SearchUsers(ctx context.Context, req *pb.SearchUsersRequest) (*pb.SearchUsersResponse, error) {
	result, err := s.service.SearchUsers(ctx, mapSearchUsersRequestToDomain(req))

	if err != nil {
		return nil, mapServiceError(err)
	}

	return &pb.SearchUsersResponse{
		Users:   mapUsersSummariesToProto(result.Users),
		HasMore: result.HasMore,
	}, nil
}

func (s *UserGRPCServer) UserExists(ctx context.Context, req *pb.UserExistsRequest) (*pb.UserExistsResponse, error) {
	lookup, err := mapUserExistsRequestLookupToDomain(req)

	if err != nil {
		return nil, mapServiceError(err)
	}

	exists, err := s.service.UserExists(ctx, lookup)

	if err != nil {
		return nil, mapServiceError(err)
	}

	return &pb.UserExistsResponse{
		Exists: exists,
	}, nil
}

func (s *UserGRPCServer) UsersExist(ctx context.Context, req *pb.UsersExistRequest) (*pb.UsersExistResponse, error) {
	result, err := s.service.UsersExist(ctx, req.UserIds)

	if err != nil {
		return nil, mapServiceError(err)
	}

	return &pb.UsersExistResponse{
		Users: mapUserExistanceToProto(result),
	}, nil
}

func (s *UserGRPCServer) GetPrivacySettings(ctx context.Context, req *pb.GetPrivacySettingsRequest) (*pb.GetPrivacySettingsResponse, error) {
	settings, err := s.service.GetPrivacySettings(ctx, req.UserId)

	if err != nil {
		return nil, mapServiceError(err)
	}

	return &pb.GetPrivacySettingsResponse{
		Settings: mapPrivacySettingsToProto(settings),
	}, nil
}

func (s *UserGRPCServer) UpdatePrivacySettings(ctx context.Context, req *pb.UpdatePrivacySettingsRequest) (*pb.UpdatePrivacySettingsResponse, error) {
	settings := mapUpdatePrivacySettingsRequestToDomain(req)

	err := s.service.UpdatePrivacySettings(ctx, settings)

	if err != nil {
		return nil, mapServiceError(err)
	}

	return &pb.UpdatePrivacySettingsResponse{}, nil
}
