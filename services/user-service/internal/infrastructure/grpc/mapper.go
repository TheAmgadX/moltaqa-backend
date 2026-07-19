package grpc

import (
	"time"

	pb "github.com/TheAmgadX/moltaqa-backend/shared/proto/users"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/domain"
)

func mapCreateUserRequestToDomain(req *pb.CreateUserRequest) *domain.User {
	return &domain.User{
		Username:       req.Username,
		Email:          req.Email,
		PhoneNumber:    req.Phone,
		DisplayName:    req.DisplayName,
		EmailVerified:  time.Time{},
		PhoneVerified:  time.Time{},
		BirthDate:      time.Time{},
		BioStatus:      "",
		AccountBadge:   domain.UNVERIFIED,
		FriendsCount:   0,
		FollowersCount: 0,
		FollowingCount: 0,
		PostsCount:     0,
		DeletedAt:      time.Time{},
	}
}

func mapRegisterContactRequestToDomain(req *pb.RegisterContactRequest) *domain.ContactRequest {
	var contactType domain.ContactLookupType
	var value string

	switch lookup := req.ContactType.(type) {
	case *pb.RegisterContactRequest_Email:
		contactType = domain.ContactLookupTypeEmail
		value = lookup.Email
	case *pb.RegisterContactRequest_Phone:
		contactType = domain.ContactLookupTypePhone
		value = lookup.Phone
	}

	return &domain.ContactRequest{
		UserId: req.UserId,
		ContactLookup: domain.ContactLookup{
			Type:  contactType,
			Value: value,
		},
	}
}

func mapUpdateUserRequestToDomain(req *pb.UpdateUserRequest) (*domain.User, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, domain.ErrInvalidUserId
	}

	return &domain.User{
		Id:             id,
		Username:       req.Username,
		DisplayName:    req.DisplayName,
		BirthDate:      mapTimeToDomain(req.BirthDate),
		BioStatus:      req.BioStatus,
		AccountBadge:   mapAccountBadgeToDomain(req.AccountBadge),
		FriendsCount:   req.FriendsCount,
		FollowersCount: req.FollowersCount,
		FollowingCount: req.FollowingCount,
		PostsCount:     req.PostsCount,
	}, nil
}

func mapAccountBadgeToProto(accountBadge domain.AccountBadgeType) pb.AccountBadge {
	switch accountBadge {
	case domain.UNVERIFIED:
		return pb.AccountBadge_UNVERIFIED
	case domain.BLUE_BADGE:
		return pb.AccountBadge_BLUE_BADGE
	case domain.GOLDEN_BADGE:
		return pb.AccountBadge_GOLD_BADGE
	case domain.SILVER_BADGE:
		return pb.AccountBadge_SILVER_BADGE
	default:
		return pb.AccountBadge_UNVERIFIED
	}
}

func mapAccountBadgeToDomain(accountBadge pb.AccountBadge) domain.AccountBadgeType {
	switch accountBadge {
	case pb.AccountBadge_UNVERIFIED:
		return domain.UNVERIFIED
	case pb.AccountBadge_BLUE_BADGE:
		return domain.BLUE_BADGE
	case pb.AccountBadge_GOLD_BADGE:
		return domain.GOLDEN_BADGE
	case pb.AccountBadge_SILVER_BADGE:
		return domain.SILVER_BADGE
	default:
		return domain.UNVERIFIED
	}
}

func mapTimeToProto(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func mapTimeToDomain(t *timestamppb.Timestamp) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.AsTime()
}

func mapUserToProto(user *domain.User) *pb.User {
	return &pb.User{
		Id:             user.Id.String(),
		Username:       user.Username,
		Email:          user.Email,
		Phone:          user.PhoneNumber,
		DisplayName:    user.DisplayName,
		EmailVerified:  mapTimeToProto(user.EmailVerified),
		PhoneVerified:  mapTimeToProto(user.PhoneVerified),
		BirthDate:      mapTimeToProto(user.BirthDate),
		BioStatus:      user.BioStatus,
		AccountBadge:   mapAccountBadgeToProto(user.AccountBadge),
		FriendsCount:   user.FriendsCount,
		FollowersCount: user.FollowersCount,
		FollowingCount: user.FollowingCount,
		PostsCount:     user.PostsCount,
		CreatedAt:      mapTimeToProto(user.CreatedAt),
		UpdatedAt:      mapTimeToProto(user.UpdatedAt),
	}
}

func mapUsersToProto(users []domain.User) []*pb.User {
	result := make([]*pb.User, 0, len(users))

	for _, user := range users {
		result = append(result, mapUserToProto(&user))
	}

	return result
}

func mapUserSummaryToProto(userSummary *domain.UserSummary) *pb.UserSummary {
	return &pb.UserSummary{
		Id:              userSummary.Id,
		Username:        userSummary.Username,
		DisplayName:     userSummary.DisplayName,
		PhoneNumber:     userSummary.PhoneNumber,
		ProfileImageUrl: userSummary.ProfileImageURL,
		AccountBadge:    mapAccountBadgeToProto(userSummary.AccountBadge),
	}
}

func mapUsersSummaryToProto(usersSummary []domain.UserSummary) []*pb.UserSummary {
	result := make([]*pb.UserSummary, 0, len(usersSummary))

	for _, userSummary := range usersSummary {
		result = append(result, mapUserSummaryToProto(&userSummary))
	}

	return result
}

func mapGetUserRequestLookupToDomain(req *pb.GetUserRequest) (domain.Lookup, error) {
	switch lookup := req.Lookup.(type) {

	case *pb.GetUserRequest_Id:
		return domain.Lookup{
			Type:  domain.LookUpId,
			Value: lookup.Id,
		}, nil

	case *pb.GetUserRequest_Username:
		return domain.Lookup{
			Type:  domain.LookupUsername,
			Value: lookup.Username,
		}, nil

	case *pb.GetUserRequest_Email:
		return domain.Lookup{
			Type:  domain.LookupEmail,
			Value: lookup.Email,
		}, nil

	case *pb.GetUserRequest_Phone:
		return domain.Lookup{
			Type:  domain.LookupPhone,
			Value: lookup.Phone,
		}, nil

	default:
		return domain.Lookup{}, domain.ErrInvalidUserInput
	}
}

func mapUserExistsRequestLookupToDomain(req *pb.UserExistsRequest) (domain.Lookup, error) {
	switch lookup := req.Lookup.(type) {

	case *pb.UserExistsRequest_Id:
		return domain.Lookup{
			Type:  domain.LookUpId,
			Value: lookup.Id,
		}, nil

	case *pb.UserExistsRequest_Username:
		return domain.Lookup{
			Type:  domain.LookupUsername,
			Value: lookup.Username,
		}, nil

	case *pb.UserExistsRequest_Email:
		return domain.Lookup{
			Type:  domain.LookupEmail,
			Value: lookup.Email,
		}, nil

	case *pb.UserExistsRequest_Phone:
		return domain.Lookup{
			Type:  domain.LookupPhone,
			Value: lookup.Phone,
		}, nil

	default:
		return domain.Lookup{}, domain.ErrInvalidUserInput
	}
}

func mapSearchUsersRequestToDomain(req *pb.SearchUsersRequest) *domain.UserSearch {
	return &domain.UserSearch{
		Query:    req.Query,
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	}
}

func mapUsersSummariesToProto(users []domain.UserSummary) []*pb.UserSummary {
	result := make([]*pb.UserSummary, 0, len(users))

	for _, user := range users {
		result = append(result, mapUserSummaryToProto(&user))
	}

	return result
}

func mapUserExistanceToProto(result []domain.UserExistence) []*pb.UserExistence {
	protoResult := make([]*pb.UserExistence, 0, len(result))

	for _, user := range result {
		protoResult = append(protoResult, &pb.UserExistence{
			UserId: user.Id,
			Exists: user.Exists,
		})
	}

	return protoResult
}

func mapPrivacySettingsToProto(settings *domain.PrivacySettings) *pb.PrivacySettings {
	return &pb.PrivacySettings{
		AvatarVisibility:    pb.Visibility(settings.AvatarVisibility),
		PhoneVisibility:     pb.Visibility(settings.PhoneVisibility),
		EmailVisibility:     pb.Visibility(settings.EmailVisibility),
		LastSeenVisibility:  pb.Visibility(settings.LastSeenVisibility),
		ReadReceiptsEnabled: settings.ReadReceiptsEnabled,
		FindByUsername:      settings.FindByUsername,
	}
}

func mapUpdatePrivacySettingsRequestToDomain(req *pb.UpdatePrivacySettingsRequest) *domain.PrivacySettings {
	return &domain.PrivacySettings{
		AvatarVisibility:    domain.Visibility(req.AvatarVisibility.Value),
		PhoneVisibility:     domain.Visibility(req.PhoneVisibility.Value),
		EmailVisibility:     domain.Visibility(req.EmailVisibility.Value),
		LastSeenVisibility:  domain.Visibility(req.LastSeenVisibility.Value),
		ReadReceiptsEnabled: req.ReadReceiptsEnabled,
		FindByUsername:      req.FindByUsername,
	}
}
