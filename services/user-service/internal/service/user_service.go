package service

import (
	"context"

	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/domain"
)

type UserService struct {
}

func NewService() *UserService {
	return &UserService{}
}

func (s *UserService) Create(ctx context.Context, user *domain.User) error {
	return nil
}

func (s *UserService) RegisterContact(ctx context.Context, contact *domain.ContactRequest) error {
	return nil
}

func (s *UserService) Update(ctx context.Context, user *domain.User) error {
	return nil
}

func (s *UserService) VerifyEmail(ctx context.Context, userId string) error {
	return nil
}

func (s *UserService) VerifyPhone(ctx context.Context, userId string) error {
	return nil
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	return nil
}

func (s *UserService) Restore(ctx context.Context, id string) error {
	return nil
}

func (s *UserService) Get(ctx context.Context, lookup domain.Lookup) (*domain.User, error) {
	return nil, nil
}

func (s *UserService) GetUsers(ctx context.Context, userIds []string) ([]domain.User, error) {
	return nil, nil
}

func (s *UserService) GetUserSummary(ctx context.Context, id string) (*domain.UserSummary, error) {
	return nil, nil
}

func (s *UserService) GetUsersSummary(ctx context.Context, userIds []string) ([]domain.UserSummary, error) {
	return nil, nil
}

func (s *UserService) SearchUsers(ctx context.Context, req *domain.UserSearch) (*domain.UserSearchResult, error) {
	return nil, nil
}

func (s *UserService) UserExists(ctx context.Context, lookup domain.Lookup) (bool, error) {
	return false, nil
}

func (s *UserService) UsersExist(ctx context.Context, userIds []string) ([]domain.UserExistence, error) {
	return nil, nil
}

func (s *UserService) GetPrivacySettings(ctx context.Context, userId string) (*domain.PrivacySettings, error) {
	return nil, nil
}

func (s *UserService) UpdatePrivacySettings(ctx context.Context, settings *domain.PrivacySettings) error {
	return nil
}
