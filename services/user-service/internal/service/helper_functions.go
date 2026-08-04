package service

import (
	"context"
	"time"

	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/domain"
	events "github.com/TheAmgadX/moltaqa-backend/shared/events/users"
	"github.com/TheAmgadX/moltaqa-backend/shared/kafka"
	"github.com/google/uuid"
)

// fillUserDefaultData sets default values for a user.
//
// This function is used in the user creation process.
// It depends on the grpc mapper to set the phone and email.
func fillUserDefaultData(user *domain.User) {
	user.Id = uuid.New()
	user.Username = "user_" + uuid.New().String()
	user.DisplayName = "New User"
	user.BirthDate = time.Time{}
	user.Bio = ""
	user.BioStatus = ""
}

func (s *UserService) publishUpdateUserEvents(ctx context.Context, user *domain.UserUpdate) {
	if user.Bio != nil {
		event := &events.UserUpdated{
			UserId: user.Id,
			Field:  "bio",
			Value:  *user.Bio,
		}

		s.ProduceEvent(ctx, kafka.UserUpdated, user.Id, event)
	}

	if user.BioStatus != nil {
		event := &events.UserUpdated{
			UserId: user.Id,
			Field:  "bio_status",
			Value:  *user.BioStatus,
		}

		s.ProduceEvent(ctx, kafka.UserUpdated, user.Id, event)
	}

	if user.DisplayName != nil {
		event := &events.UserUpdated{
			UserId: user.Id,
			Field:  "display_name",
			Value:  *user.DisplayName,
		}

		s.ProduceEvent(ctx, kafka.UserUpdated, user.Id, event)
	}

	if user.Username != nil {
		event := &events.UserUpdated{
			UserId: user.Id,
			Field:  "username",
			Value:  *user.Username,
		}

		s.ProduceEvent(ctx, kafka.UserUpdated, user.Id, event)
	}

	if user.ProfileImageUrl != nil {
		event := &events.UserUpdated{
			UserId: user.Id,
			Field:  "profile_image",
			Value:  *user.ProfileImageUrl,
		}

		s.ProduceEvent(ctx, kafka.UserUpdated, user.Id, event)
	}

	if user.BirthDate != nil {
		event := &events.UserUpdated{
			UserId: user.Id,
			Field:  "birth_date",
			Value:  user.BirthDate.String(),
		}

		s.ProduceEvent(ctx, kafka.UserUpdated, user.Id, event)
	}
}
