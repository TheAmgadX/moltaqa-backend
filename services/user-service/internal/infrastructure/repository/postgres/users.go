package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/domain"
	utils_postgres "github.com/TheAmgadX/moltaqa-backend/shared/utils/postgres"
	"github.com/jackc/pgx/v5"
)

func (r *UserPostgresRepository) Create(ctx context.Context, user *domain.User) error {
	if user == nil {
		return domain.ErrInvalidUserInput
	}

	// begin a transaction
	// the user creation should not be committed until the privacy settings are also created
	// to make sure the user and privacy settings are both created or neither are created
	tx, err := r.db.Begin(ctx)

	if err != nil {
		return utils_postgres.MapDBErrorToServiceError(err)
	}

	// make sure to rollback the transaction if an error occurs
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	query := `
		INSERT INTO users (
		    id,
		    username,
		    email,
		    phone,
		    display_name,
		    birth_date,
		    bio,
		    bio_status
		)
		VALUES (
		    $1,
		    $2,
		    $3,
		    $4,
		    $5,
		    $6,
		    $7,
		    $8
		);
	`

	_, err = r.db.Exec(
		ctx,
		query,
		user.Id,
		user.Username,
		user.Email,
		user.PhoneNumber,
		user.DisplayName,
		user.BirthDate,
		user.Bio,
		user.BioStatus,
	)

	if err != nil {
		return utils_postgres.MapDBErrorToServiceError(err)
	}

	// add the privacy settings row with default values
	query = `
		INSERT INTO privacy_settings (
			user_id,
		)
		VALUES (
		    $1
		)
	`

	_, err = r.db.Exec(
		ctx,
		query,
		user.Id,
	)

	if err != nil {
		return utils_postgres.MapDBErrorToServiceError(err)
	}

	err = tx.Commit(ctx)

	return utils_postgres.MapDBErrorToServiceError(err)
}

func (r *UserPostgresRepository) Update(ctx context.Context, userUpdate *domain.UserUpdate) error {
	// dynamic SQL generation
	var (
		sets []string
		args []any
		i    = 1
	)

	if userUpdate.Username != nil {
		sets = append(sets, fmt.Sprintf("username = $%d", i))
		args = append(args, *userUpdate.Username)
		i++
	}

	if userUpdate.DisplayName != nil {
		sets = append(sets, fmt.Sprintf("display_name = $%d", i))
		args = append(args, *userUpdate.DisplayName)
		i++
	}

	if userUpdate.Email != nil {
		sets = append(sets, fmt.Sprintf("email = $%d", i))
		args = append(args, *userUpdate.Email)
		i++
	}

	if userUpdate.Phone != nil {
		sets = append(sets, fmt.Sprintf("phone = $%d", i))
		args = append(args, *userUpdate.Phone)
		i++
	}

	if userUpdate.ProfileImageURL != nil {
		sets = append(sets, fmt.Sprintf("profile_image_url = $%d", i))
		args = append(args, *userUpdate.ProfileImageURL)
		i++
	}

	if userUpdate.Bio != nil {
		sets = append(sets, fmt.Sprintf("bio = $%d", i))
		args = append(args, *userUpdate.Bio)
		i++
	}

	if userUpdate.BioStatus != nil {
		sets = append(sets, fmt.Sprintf("bio_status = $%d", i))
		args = append(args, *userUpdate.BioStatus)
		i++
	}

	if userUpdate.BirthDate != nil {
		sets = append(sets, fmt.Sprintf("birth_date = $%d", i))
		args = append(args, *userUpdate.BirthDate)
		i++
	}

	if len(sets) == 0 {
		return domain.ErrNothingToUpdate
	}

	sets = append(sets, "updated_at = NOW()")

	// build the query
	// id is the last passed argument so it's passed as i (the counter.)
	query := fmt.Sprintf(`
		UPDATE users
		SET %s
		WHERE id = $%d
	`, strings.Join(sets, ", "), i)

	args = append(args, userUpdate.Id)

	_, err := r.db.Exec(ctx, query, args...)

	return utils_postgres.MapDBErrorToServiceError(err)
}

func (r *UserPostgresRepository) SoftDelete(ctx context.Context, id string) error {
	query := `
		UPDATE users
		SET soft_delete = NOW()
		WHERE id = $1;
	`

	if id == "" {
		return domain.ErrInvalidUserId
	}

	_, err := r.db.Exec(ctx, query, id)

	return utils_postgres.MapDBErrorToServiceError(err)
}

func mapDBRowToUser(rows *pgx.Rows) (*domain.User, error) {
	var user domain.User

	if (*rows).Next() {
		err := (*rows).Scan(&user.Id, &user.Username, &user.PhoneNumber,
			&user.Email, &user.ProfileImageUrl, &user.Bio, &user.DisplayName,
			&user.EmailVerified, &user.PhoneVerified, &user.BirthDate,
			&user.BioStatus, &user.AccountBadge, &user.FriendsCount,
			&user.FollowersCount, &user.FollowingCount, &user.PostsCount,
			&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)

		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (r *UserPostgresRepository) Get(ctx context.Context, lookup domain.Lookup) (*domain.User, error) {
	query := fmt.Sprintf(`
		SELECT * FROM users
		WHERE %s = $1
	`, lookup.TypeString())

	rows, err := r.db.Query(ctx, query, lookup.Value)

	if err != nil {
		return nil, utils_postgres.MapDBErrorToServiceError(err)
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, domain.ErrUserNotFound
	}

	user, err := mapDBRowToUser(&rows)

	if err != nil {
		return nil, utils_postgres.MapDBErrorToServiceError(err)
	}

	return user, nil
}

func (r *UserPostgresRepository) GetUsers(ctx context.Context, ids []string) ([]domain.User, error) {
	if len(ids) == 0 {
		return nil, domain.ErrEmptyUserIdSlice
	}

	query := `
			SELECT *
			FROM users
			WHERE id = ANY($1)
		`

	rows, err := r.db.Query(ctx, query, ids)

	if err != nil {
		return nil, utils_postgres.MapDBErrorToServiceError(err)
	}
	defer rows.Close()

	var users []domain.User

	for rows.Next() {
		user, err := mapDBRowToUser(&rows)
		if err != nil {
			return nil, utils_postgres.MapDBErrorToServiceError(err)
		}

		users = append(users, *user)
	}

	if err := rows.Err(); err != nil {
		return nil, utils_postgres.MapDBErrorToServiceError(err)
	}

	return users, nil
}

// getUserSummaryFieldsString returns the users summary fields as string
// used to build the SQL query for getting user summaries.
//
// warning: this should be updated whenever the user summary fields change.
func getUserSummaryFieldsString() []string {
	return []string{
		"id", "username", "display_name", "phone_number", "profile_image_url", "profile_badge",
	}
}

func mapDBRowToUserSummary(rows *pgx.Rows) (*domain.UserSummary, error) {
	var user domain.UserSummary

	if err := (*rows).Scan(&user.Id, &user.Username, &user.DisplayName,
		&user.PhoneNumber, &user.ProfileImageURL, &user.ProfileBadge); err != nil {
		return nil, utils_postgres.MapDBErrorToServiceError(err)
	}

	return &user, nil
}

func (r *UserPostgresRepository) GetSummary(ctx context.Context, id string) (*domain.UserSummary, error) {
	if id == "" {
		return nil, domain.ErrInvalidUserId
	}

	query := `
		SELECT ` + strings.Join(getUserSummaryFieldsString(), ", ") + `
		FROM users
		WHERE id = $1
	`

	rows, err := r.db.Query(ctx, query, id)

	if err != nil {
		return nil, utils_postgres.MapDBErrorToServiceError(err)
	}

	defer rows.Close()

	if !rows.Next() {
		return nil, domain.ErrUserNotFound
	}

	user, err := mapDBRowToUserSummary(&rows)

	if err != nil {
		return nil, utils_postgres.MapDBErrorToServiceError(err)
	}

	return user, nil
}

func (r *UserPostgresRepository) GetSummaries(ctx context.Context, ids []string) ([]domain.UserSummary, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := `
		SELECT ` + strings.Join(getUserSummaryFieldsString(), ", ") + `
		FROM users
		WHERE id ANY ($1)
	`

	rows, err := r.db.Query(ctx, query, ids)

	if err != nil {
		return nil, utils_postgres.MapDBErrorToServiceError(err)
	}
	defer rows.Close()

	var users []domain.UserSummary

	for rows.Next() {
		user, err := mapDBRowToUserSummary(&rows)

		if err != nil {
			return nil, utils_postgres.MapDBErrorToServiceError(err)
		}

		users = append(users, *user)
	}

	return users, nil
}

// Search searches for users based on the given query, page, and page size.
//
// query is a string to be searched for the following:
//
// - username
//
// - display name
//
// Search doesn't support searching based on others fields.
// This search functionality gonna be replaced by an intelligent search service.
func (r *UserPostgresRepository) Search(ctx context.Context, query string, page, pageSize uint32) (*domain.UserSearchResult, error) {
	offset := (page - 1) * pageSize
	search := "%" + query + "%"

	// Count total matches
	// TODO: Optimize this extra query:
	// I am thinking in two options:
	// 1. Changing the repo api to support caching for the result of this query.
	// 2. Quering for offset + 1 and change the grpc api to support
	//	  having a field of `has_next` to indicate if there are more results.
	// I prefer option 2.
	var total uint32
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM users u
		JOIN privacy_settings ps
		    ON ps.user_id = u.id
		WHERE ps.find_by_username = TRUE
		  AND (
		      u.username ILIKE $1
		      OR u.display_name ILIKE $1
  );
	`, search).Scan(&total)

	if err != nil {
		return nil, utils_postgres.MapDBErrorToServiceError(err)
	}

	// Second query to fetch paginated results
	rows, err := r.db.Query(ctx, `
		SELECT u.*
		FROM users u
		JOIN privacy_settings ps
		    ON ps.user_id = u.id
		WHERE ps.find_by_username = TRUE
		  AND (
		      u.username ILIKE $1
		      OR u.display_name ILIKE $1
		  )
		ORDER BY u.username
		LIMIT $2
		OFFSET $3;
	`, search, pageSize, offset)

	if err != nil {
		return nil, utils_postgres.MapDBErrorToServiceError(err)
	}
	defer rows.Close()

	users := make([]domain.UserSummary, 0)

	for rows.Next() {
		user, err := mapDBRowToUserSummary(&rows)
		if err != nil {
			return nil, err
		}

		users = append(users, *user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &domain.UserSearchResult{
		Users:        users,
		TotalResults: total,
	}, nil
}

// Validation
func (r *UserPostgresRepository) Exists(ctx context.Context, lookup domain.Lookup) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE ` + lookup.TypeString() + ` = $1
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, lookup.Value).Scan(&exists)

	// default return value by .Scan if no values where found.
	if err == pgx.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, utils_postgres.MapDBErrorToServiceError(err)
	}

	return exists, nil
}

func (r *UserPostgresRepository) UsersExist(ctx context.Context, ids []string) ([]domain.UserExistence, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := `
		SELECT id, EXISTS (
			SELECT 1
			FROM users
			WHERE id = ANY ($1)
		)
	`

	rows, err := r.db.Query(ctx, query, ids)

	if err != nil {
		return nil, utils_postgres.MapDBErrorToServiceError(err)
	}
	defer rows.Close()

	var users []domain.UserExistence

	for rows.Next() {
		var id string
		var exists bool
		err := rows.Scan(&id, &exists)

		if err != nil {
			return nil, utils_postgres.MapDBErrorToServiceError(err)
		}

		users = append(users, domain.NewUserExistence(id, exists))
	}

	return users, nil
}
