package repository

import (
	"context"

	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/domain"
)

// UserRepository defines the persistence layer for user-related data.
//
// Collections are returned as slices of values rather than pointers.
// This improves cache locality and avoids unnecessary heap allocations
// for read-only query models such as UserSummary.
type UserRepository interface {
	// Commands
	Create(ctx context.Context, user *domain.User) error
	RegisterContact(ctx context.Context, user *domain.ContactRequest) error
	Update(ctx context.Context, userUpdate *domain.UserUpdate) error
	VerifyEmail(ctx context.Context, id string) error
	VerifyPhone(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id string) error
	RestoreUser(ctx context.Context, id string) error

	// Queries
	Get(ctx context.Context, lookup domain.Lookup) (*domain.User, error)
	GetUsers(ctx context.Context, ids []string) ([]domain.User, error)
	GetSummary(ctx context.Context, id string) (*domain.UserSummary, error)
	GetSummaries(ctx context.Context, ids []string) ([]domain.UserSummary, error)
	Search(ctx context.Context, query string, page, pageSize uint32) (*domain.UserSearchResult, error)

	// Validation
	Exists(ctx context.Context, lookup domain.Lookup) (bool, error)
	UsersExist(ctx context.Context, ids []string) ([]domain.UserExistence, error)

	// Privacy Settings
	GetPrivacySettings(ctx context.Context, id string) (*domain.PrivacySettings, error)
	UpdatePrivacySettings(ctx context.Context, id string, settingsUpdate *domain.PrivacySettingsUpdate) error
}
