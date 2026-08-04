package grpc

import (
	"time"

	pb "github.com/TheAmgadX/moltaqa-backend/shared/proto/users"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/domain"
)

func mapCreateUserRequestToDomain(req *pb.CreateUserRequest) *domain.User {

	email := ""
	phone := ""

	switch lookup := req.Contact.(type) {
	case *pb.CreateUserRequest_Email:
		email = lookup.Email
	case *pb.CreateUserRequest_Phone:
		phone = lookup.Phone
	}

	return &domain.User{
		Email:       email,
		PhoneNumber: phone,
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

func mapUpdateUserRequestToDomain(req *pb.UpdateUserRequest) (*domain.UserUpdate, error) {
	user := &domain.UserUpdate{
		Id:              req.Id,
		Username:        req.Username,
		ProfileImageUrl: req.ProfileImageUrl,
		DisplayName:     req.DisplayName,
		Bio:             req.Bio,
		BioStatus:       req.BioStatus,
		FriendsCount:    req.FriendsCount,
		FollowersCount:  req.FollowersCount,
		FollowingCount:  req.FollowingCount,
		PostsCount:      req.PostsCount,
	}

	if req.BirthDate != nil {
		t := mapTimeToDomain(req.BirthDate)
		user.BirthDate = &t
	}

	if req.AccountBadge != nil {
		badge := mapAccountBadgeToDomain(*req.AccountBadge)
		user.AccountBadge = &badge
	}

	return user, nil
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

func mapUserExistanceToProto(user domain.UserExistence) *pb.UserExistence {
	return &pb.UserExistence{
		UserId: user.Id,
		Exists: user.Exists,
	}
}

func mapUsersExistanceToProto(result []domain.UserExistence) []*pb.UserExistence {
	protoResult := make([]*pb.UserExistence, 0, len(result))

	for _, user := range result {
		protoResult = append(protoResult, mapUserExistanceToProto(user))
	}

	return protoResult
}

func mapVisibilityToProto(visibility domain.Visibility) pb.Visibility {
	switch visibility {
	case domain.EVERYONE:
		return pb.Visibility_EVERYONE
	case domain.FRIENDS:
		return pb.Visibility_FRIENDS
	case domain.CONTACTS:
		return pb.Visibility_CONTACTS
	case domain.NOBODY:
		return pb.Visibility_NOBODY
	default:
		return pb.Visibility_NOBODY
	}
}

func mapPrivacySettingsToProto(settings *domain.PrivacySettings) *pb.PrivacySettings {
	return &pb.PrivacySettings{
		AvatarVisibility:    mapVisibilityToProto(settings.AvatarVisibility),
		PhoneVisibility:     mapVisibilityToProto(settings.PhoneVisibility),
		EmailVisibility:     mapVisibilityToProto(settings.EmailVisibility),
		LastSeenVisibility:  mapVisibilityToProto(settings.LastSeenVisibility),
		ReadReceiptsEnabled: settings.ReadReceiptsEnabled,
		FindByUsername:      settings.FindByUsername,
	}
}

func MapProtoVisibilityToDomain(val pb.Visibility) domain.Visibility {
	switch val {
	case pb.Visibility_EVERYONE:
		return domain.EVERYONE
	case pb.Visibility_FRIENDS:
		return domain.FRIENDS
	case pb.Visibility_CONTACTS:
		return domain.CONTACTS
	case pb.Visibility_NOBODY:
		return domain.NOBODY
	default:
		// Catches invalid integers like 999 sent over gRPC
		return ""
	}
}

func mapUpdatePrivacySettingsRequestToDomain(req *pb.UpdatePrivacySettingsRequest) *domain.PrivacySettingsUpdate {
	settings := &domain.PrivacySettingsUpdate{}

	settings.UserId = req.UserId

	if req.AvatarVisibility != nil {
		vis := MapProtoVisibilityToDomain(*req.AvatarVisibility)
		settings.AvatarVisibility = &vis
	}

	if req.PhoneVisibility != nil {
		vis := MapProtoVisibilityToDomain(*req.PhoneVisibility)
		settings.PhoneVisibility = &vis
	}

	if req.EmailVisibility != nil {
		vis := MapProtoVisibilityToDomain(*req.EmailVisibility)
		settings.EmailVisibility = &vis
	}

	if req.LastSeenVisibility != nil {
		vis := MapProtoVisibilityToDomain(*req.LastSeenVisibility)
		settings.LastSeenVisibility = &vis
	}

	// 2. Booleans
	if req.ReadReceiptsEnabled != nil {
		settings.ReadReceiptsEnabled = req.ReadReceiptsEnabled
	}

	if req.FindByUsername != nil {
		settings.FindByUsername = req.FindByUsername
	}

	return settings
}
