package service

import (
	"context"

	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/domain"
	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/infrastructure/repository"
	events "github.com/TheAmgadX/moltaqa-backend/shared/events/users"
	"github.com/TheAmgadX/moltaqa-backend/shared/kafka"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

type UserService struct {
	repo     repository.UserRepository
	producer *kafka.Producer
}

func NewService(repo repository.UserRepository, client *kgo.Client) (*UserService, error) {
	return &UserService{repo: repo, producer: kafka.NewProducer(client)}, nil
}

func (s *UserService) ProduceEvent(ctx context.Context, topic kafka.Topic, key string, event proto.Message) {
	err := s.producer.Produce(ctx, topic, key, event)

	if err != nil {
		// TODO: log here using the logger when implement it.
	}
}

// Create creates a new user
//
// Parameters:
// - ctx: the context for the request
// - user: the user to create
//
// Creating users in the process is just by adding default values for the essential feilds except the login credentials
// The login credentials are provided by gRPC and are not set by this method
//
// Publishes UserRegistered event when created.
// gRPC Response with the user_id of the created user
func (s *UserService) Create(ctx context.Context, user *domain.User) error {
	if err := createUserValidation(user); err != nil {
		return err
	}

	fillUserDefaultData(user)

	err := s.repo.Create(ctx, user)

	if err != nil {
		return domain.MapPostgresErrorToDomain(err)
	}

	event := &events.UserRegistered{
		Id: user.Id.String(),
	}

	s.ProduceEvent(ctx, kafka.UserRegistered, user.Id.String(), event)

	return nil
}

func (s *UserService) RegisterContact(ctx context.Context, contact *domain.ContactRequest) error {
	if err := registerContactValidation(contact); err != nil {
		return err
	}

	err := s.repo.RegisterContact(ctx, contact)
	if err != nil {
		return domain.MapPostgresErrorToDomain(err)
	}

	event := &events.ContactRegistered{
		UserId: contact.UserId,
	}

	switch contact.TypeString() {
	case "email":
		event.Contact = &events.ContactRegistered_Email{
			Email: contact.Value(),
		}
	case "phone":
		event.Contact = &events.ContactRegistered_Phone{
			Phone: contact.Value(),
		}
	}

	s.ProduceEvent(ctx, kafka.ContactRegistered, contact.UserId, event)

	return nil
}

func (s *UserService) Update(ctx context.Context, user *domain.UserUpdate) error {
	if err := updateUserValidation(user); err != nil {
		return err
	}

	err := s.repo.Update(ctx, user)
	if err != nil {
		return domain.MapPostgresErrorToDomain(err)
	}

	s.publishUpdateUserEvents(ctx, user)
	return nil
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	err := s.repo.SoftDelete(ctx, id)
	if err != nil {
		return domain.MapPostgresErrorToDomain(err)
	}

	event := &events.UserDeleted{}

	s.ProduceEvent(ctx, kafka.UserDeleted, id, event)

	return nil
}

func (s *UserService) Restore(ctx context.Context, id string) error {
	err := s.repo.RestoreUser(ctx, id)

	if err != nil {
		return domain.MapPostgresErrorToDomain(err)
	}

	event := &events.UserRestored{}

	s.ProduceEvent(ctx, kafka.UserRestored, id, event)

	return nil
}

func (s *UserService) Get(ctx context.Context, lookup domain.Lookup) (*domain.User, error) {
	if err := lookupValidation(lookup); err != nil {
		return nil, err
	}

	user, err := s.repo.Get(ctx, lookup)

	if err != nil {
		return nil, domain.MapPostgresErrorToDomain(err)
	}

	return user, nil
}

func (s *UserService) GetUsers(ctx context.Context, userIds []string) ([]domain.User, error) {
	if len(userIds) == 0 {
		return nil, nil
	}

	users, err := s.repo.GetUsers(ctx, userIds)
	if err != nil {
		return nil, domain.MapPostgresErrorToDomain(err)
	}

	return users, nil
}

func (s *UserService) GetUserSummary(ctx context.Context, id string) (*domain.UserSummary, error) {
	if id == "" {
		return nil, domain.NewValidationError("invalid user id.")
	}

	user, err := s.repo.GetSummary(ctx, id)
	if err != nil {
		return nil, domain.MapPostgresErrorToDomain(err)
	}

	return user, nil
}

func (s *UserService) GetUsersSummary(ctx context.Context, userIds []string) ([]domain.UserSummary, error) {
	if len(userIds) == 0 {
		return nil, nil
	}

	users, err := s.repo.GetSummaries(ctx, userIds)
	if err != nil {
		return nil, domain.MapPostgresErrorToDomain(err)
	}

	return users, nil
}

func (s *UserService) SearchUsers(ctx context.Context, req *domain.UserSearch) (*domain.UserSearchResult, error) {
	// TODO: implement validation for the request.
	if req == nil {
		return nil, nil
	}

	users, err := s.repo.Search(ctx, req)
	if err != nil {
		return nil, domain.MapPostgresErrorToDomain(err)
	}

	return users, nil
}

func (s *UserService) UserExists(ctx context.Context, lookup domain.Lookup) (domain.UserExistence, error) {
	if err := lookupValidation(lookup); err != nil {
		return domain.UserExistence{}, err
	}

	id, err := s.repo.Exists(ctx, lookup)
	if err != nil {
		return domain.UserExistence{}, domain.MapPostgresErrorToDomain(err)
	}

	return domain.UserExistence{Id: id, Exists: true}, nil
}

func (s *UserService) UsersExist(ctx context.Context, userIds []string) ([]domain.UserExistence, error) {
	// TODO: implement validation for userIds
	if len(userIds) == 0 {
		return nil, nil
	}

	exists, err := s.repo.UsersExist(ctx, userIds)
	if err != nil {
		return nil, domain.MapPostgresErrorToDomain(err)
	}

	return exists, nil
}

func (s *UserService) GetPrivacySettings(ctx context.Context, userId string) (*domain.PrivacySettings, error) {
	// TODO: implement validation for userId
	if userId == "" {
		return nil, domain.NewValidationError("invalid user id.")
	}

	settings, err := s.repo.GetPrivacySettings(ctx, userId)
	if err != nil {
		return nil, domain.MapPostgresErrorToDomain(err)
	}

	return settings, nil
}

func (s *UserService) UpdatePrivacySettings(ctx context.Context, settings *domain.PrivacySettingsUpdate) error {
	// TODO: implement validation for settings
	if err := updatePrivacySettingsValidation(settings); err != nil {
		return err
	}

	if err := s.repo.UpdatePrivacySettings(ctx, settings); err != nil {
		return domain.MapPostgresErrorToDomain(err)
	}

	return nil
}
