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

	if userUpdate.PhoneNumber != nil {
		sets = append(sets, fmt.Sprintf("phone_number = $%d", i))
		args = append(args, *userUpdate.PhoneNumber)
		i++
	}

	if userUpdate.ProfileImageUrl != nil {
		sets = append(sets, fmt.Sprintf("profile_image_url = $%d", i))
		args = append(args, *userUpdate.ProfileImageUrl)
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

	if userUpdate.AccountBadge != nil {
		sets = append(sets, fmt.Sprintf("account_badge = $%d", i))
		args = append(args, *userUpdate.AccountBadge)
		i++
	}

	if userUpdate.PhoneVerified != nil {
		sets = append(sets, fmt.Sprintf("phone_verified = $%d", i))
		args = append(args, *userUpdate.PhoneVerified)
		i++
	}

	if userUpdate.EmailVerified != nil {
		sets = append(sets, fmt.Sprintf("email_verified = $%d", i))
		args = append(args, *userUpdate.EmailVerified)
		i++
	}

	if userUpdate.FriendsCount != nil {
		sets = append(sets, fmt.Sprintf("friends_count = $%d", i))
		args = append(args, *userUpdate.FriendsCount)
		i++
	}

	if userUpdate.FollowersCount != nil {
		sets = append(sets, fmt.Sprintf("followers_count = $%d", i))
		args = append(args, *userUpdate.FollowersCount)
		i++
	}

	if userUpdate.FollowingCount != nil {
		sets = append(sets, fmt.Sprintf("following_count = $%d", i))
		args = append(args, *userUpdate.FollowingCount)
		i++
	}

	if userUpdate.PostsCount != nil {
		sets = append(sets, fmt.Sprintf("posts_count = $%d", i))
		args = append(args, *userUpdate.PostsCount)
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

func (r *UserPostgresRepository) RestoreUser(ctx context.Context, id string) error {
	query := `
		UPDATE users
		SET soft_delete = NULL
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
		&user.PhoneNumber, &user.ProfileImageURL, &user.AccountBadge); err != nil {
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
func (r *UserPostgresRepository) Search(ctx context.Context, user_search *domain.UserSearch) (*domain.UserSearchResult, error) {
	offset := (user_search.Page - 1) * user_search.PageSize
	search := "%" + user_search.Query + "%"

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
	`, search, user_search.PageSize+1, offset)

	if err != nil {
		return nil, utils_postgres.MapDBErrorToServiceError(err)
	}
	defer rows.Close()

	users := make([]domain.UserSummary, 0, user_search.PageSize+1)

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

	var hasMore bool

	// if the results is more than the page size, we have more to fetch
	// remove that extra user that was fetched to check if there are more
	if len(users) > int(user_search.PageSize) {
		hasMore = true
		users = users[:user_search.PageSize]
	}

	return &domain.UserSearchResult{
		Users:   users,
		HasMore: hasMore,
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
