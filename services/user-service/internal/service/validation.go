package service

import (
	"regexp"
	"time"

	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/domain"
	"github.com/TheAmgadX/moltaqa-backend/shared/utils/assets"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/nyaruka/phonenumbers"
)

var validate = validator.New()

var validUsernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func validateEmail(email string) error {
	if err := validate.Var(email, "required,email"); err != nil {
		return domain.NewValidationError("invalid email: %s", email)
	}
	return nil
}

func validatePhone(phone string) error {
	num, err := phonenumbers.Parse(phone, "")
	if err != nil {
		return domain.NewValidationError("invalid phone number: %s", phone)
	}

	if !phonenumbers.IsValidNumber(num) {
		return domain.NewValidationError("invalid phone number: %s", phone)
	}

	return nil
}

// createUserValidation validates the user data before creating a new user.
// The request only passes the email or the phone so it validates them.
func createUserValidation(user *domain.User) error {
	if err := validateEmail(user.Email); err != nil && user.Email != "" {
		return err
	}

	if err := validatePhone(user.PhoneNumber); err != nil && user.PhoneNumber != "" {
		return err
	}

	if user.Email == "" && user.PhoneNumber == "" {
		return domain.NewValidationError("email or phone number is required")
	}

	return nil
}

// registerContactValidation validates the contact data before registering a new contact.
// The check of the user id existance id done within the update query in the repository.
func registerContactValidation(contact *domain.ContactRequest) error {
	if contact.Value() == "" {
		return domain.NewValidationError("contact is required")
	}

	if contact.TypeString() == "email" {
		if err := validateEmail(contact.Value()); err != nil {
			return err
		}
	}

	if contact.TypeString() == "phone" {
		if err := validatePhone(contact.Value()); err != nil {
			return err
		}
	}

	return nil
}

// updateUserValidation validates the user data before updating an existing user.
func updateUserValidation(user *domain.UserUpdate) error {
	_, err := uuid.Parse(user.Id)
	if err != nil {
		return domain.NewValidationError("invalid user id: %s", user.Id)
	}

	if user.Bio != nil {
		if len(*user.Bio) > 250 {
			return domain.NewValidationError("bio cannot exceed 250 characters (got %d)", len(*user.Bio))
		}
	}

	if user.BioStatus != nil {
		if len(*user.BioStatus) > 50 {
			return domain.NewValidationError("bio status cannot exceed 50 characters (got %d)", len(*user.BioStatus))
		}
	}

	if user.DisplayName != nil {
		if len(*user.DisplayName) > 50 {
			return domain.NewValidationError("display name cannot exceed 50 characters (got %d)", len(*user.DisplayName))
		}
	}

	if user.Username != nil {
		if len(*user.Username) > 50 {
			return domain.NewValidationError("username cannot exceed 50 characters (got %d)", len(*user.Username))
		}

		if !validUsernameRegex.MatchString(*user.Username) {
			return domain.NewValidationError("username can only contain english letters, numbers, and underscores.")
		}
	}

	if user.ProfileImageUrl != nil {
		if err := assets.ValidateProfileImagePath(*user.ProfileImageUrl); err != nil {
			return domain.NewValidationError("invalid profile image url: %v", err)
		}
	}

	if user.BirthDate != nil {
		if user.BirthDate.After(time.Now()) {
			return domain.NewValidationError("birth date cannot be in the future")
		}
	}

	return nil
}

func lookupValidation(lookup domain.Lookup) error {
	if lookup.TypeString() == "email" {
		if err := validateEmail(lookup.Value); err != nil {
			return err
		}
	}

	if lookup.TypeString() == "phone" {
		if err := validatePhone(lookup.Value); err != nil {
			return err
		}
	}

	return nil
}

func updatePrivacySettingsValidation(settings *domain.PrivacySettingsUpdate) error {
	if settings == nil {
		return domain.NewValidationError("invalid update privacy settings request")
	}

	if settings.AvatarVisibility != nil {
		if !settings.AvatarVisibility.IsValid() {
			return domain.NewValidationError("invalid avatar visibility: %s", *settings.AvatarVisibility)
		}
	}

	if settings.EmailVisibility != nil {
		if !settings.EmailVisibility.IsValid() {
			return domain.NewValidationError("invalid email visibility: %s", *settings.EmailVisibility)
		}
	}

	if settings.PhoneVisibility != nil {
		if !settings.PhoneVisibility.IsValid() {
			return domain.NewValidationError("invalid phone visibility: %s", *settings.PhoneVisibility)
		}
	}

	if settings.LastSeenVisibility != nil {
		if !settings.LastSeenVisibility.IsValid() {
			return domain.NewValidationError("invalid last seen visibility: %s", *settings.LastSeenVisibility)
		}
	}

	return nil
}
