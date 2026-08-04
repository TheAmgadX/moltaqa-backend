package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/domain"
	utils_postgres "github.com/TheAmgadX/moltaqa-backend/shared/utils/postgres"
	"github.com/google/uuid"
)

func (r *UserPostgresRepository) GetPrivacySettings(ctx context.Context, id string) (*domain.PrivacySettings, error) {
	query := `
        SELECT avatar_visibility, phone_visibility, email_visibility,
               last_seen_visibility, read_receipts_enabled, find_by_username
        FROM privacy_settings WHERE user_id = $1
    `
	var settings domain.PrivacySettings
	user_id, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	settings.UserId = user_id

	err = r.db.QueryRow(ctx, query, id).Scan(
		&settings.AvatarVisibility, &settings.PhoneVisibility,
		&settings.EmailVisibility, &settings.LastSeenVisibility,
		&settings.ReadReceiptsEnabled, &settings.FindByUsername,
	)

	if err != nil {
		return nil, utils_postgres.MapDBErrorToServiceError(err)
	}

	return &settings, nil
}

func (r *UserPostgresRepository) UpdatePrivacySettings(ctx context.Context, settingsUpdate *domain.PrivacySettingsUpdate) error {
	// Build the SQL query dynamically based on the fields to update
	var (
		sets []string
		args []any
		i    = 1
	)

	if settingsUpdate.AvatarVisibility != nil {
		sets = append(sets, fmt.Sprintf("avatar_visibility = $%d", i))
		args = append(args, *settingsUpdate.AvatarVisibility)
		i++
	}

	if settingsUpdate.PhoneVisibility != nil {
		sets = append(sets, fmt.Sprintf("phone_visibility = $%d", i))
		args = append(args, *settingsUpdate.PhoneVisibility)
		i++
	}

	if settingsUpdate.EmailVisibility != nil {
		sets = append(sets, fmt.Sprintf("email_visibility = $%d", i))
		args = append(args, *settingsUpdate.EmailVisibility)
		i++
	}

	if settingsUpdate.LastSeenVisibility != nil {
		sets = append(sets, fmt.Sprintf("last_seen_visibility = $%d", i))
		args = append(args, *settingsUpdate.LastSeenVisibility)
		i++
	}

	if settingsUpdate.ReadReceiptsEnabled != nil {
		sets = append(sets, fmt.Sprintf("read_receipts_enabled = $%d", i))
		args = append(args, *settingsUpdate.ReadReceiptsEnabled)
		i++
	}

	if settingsUpdate.FindByUsername != nil {
		sets = append(sets, fmt.Sprintf("find_by_username = $%d", i))
		args = append(args, *settingsUpdate.FindByUsername)
		i++
	}

	if len(sets) == 0 {
		return utils_postgres.ErrInvalidInput
	}

	// build the query
	// id is the last passed argument so it's passed as i (the counter.)
	query := fmt.Sprintf(`
		UPDATE privacy_settings
		SET %s
		WHERE user_id = $%d
	`, strings.Join(sets, ", "), i)

	args = append(args, settingsUpdate.UserId)

	_, err := r.db.Exec(ctx, query, args...)

	return utils_postgres.MapDBErrorToServiceError(err)
}
