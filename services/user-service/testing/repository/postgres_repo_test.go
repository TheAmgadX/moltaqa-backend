package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/domain"
	repository "github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/infrastructure/repository/postgres"
	utils_postgres "github.com/TheAmgadX/moltaqa-backend/shared/utils/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// ── Schema ────────────────────────────────────────────────────────────────────

const initSchema = `
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    phone_number TEXT UNIQUE NULL,
    email TEXT UNIQUE NULL,
    profile_image_url TEXT NULL,
    bio TEXT DEFAULT '',
    display_name TEXT,
    birth_date DATE NULL,
    bio_status TEXT DEFAULT '',
    account_badge TEXT NOT NULL DEFAULT 'unverified'
        CHECK (
            account_badge IN (
                'unverified',
                'blue_badge',
                'golden_badge',
                'silver_badge'
            )
        ),
    friends_count INTEGER NOT NULL DEFAULT 0,
    followers_count INTEGER NOT NULL DEFAULT 0,
    following_count INTEGER NOT NULL DEFAULT 0,
    posts_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP NULL
);

CREATE TABLE privacy_settings (
    user_id UUID PRIMARY KEY,
    avatar_visibility TEXT NOT NULL DEFAULT 'everyone'
        CHECK (avatar_visibility IN ('everyone','friends','contacts','nobody')),
    phone_visibility TEXT NOT NULL DEFAULT 'everyone'
        CHECK (phone_visibility IN ('everyone','friends','contacts','nobody')),
    email_visibility TEXT NOT NULL DEFAULT 'everyone'
        CHECK (email_visibility IN ('everyone','friends','contacts','nobody')),
    last_seen_visibility TEXT NOT NULL DEFAULT 'everyone'
        CHECK (last_seen_visibility IN ('everyone','friends','contacts','nobody')),
    read_receipts_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    find_by_username BOOLEAN NOT NULL DEFAULT TRUE,
    CONSTRAINT fk_privacy_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
`

// ── Helpers ───────────────────────────────────────────────────────────────────

// newTestRepo spins up a throwaway Postgres container, applies the schema,
// and returns a ready UserPostgresRepository along with a cleanup function.
func newTestRepo(t *testing.T) *repository.UserPostgresRepository {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	t.Cleanup(func() {
		if termErr := pgContainer.Terminate(ctx); termErr != nil {
			t.Logf("warn: failed to terminate container: %v", termErr)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Apply schema via a raw pool connection before handing off to the repo.
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to open pool for schema init: %v", err)
	}
	if _, err := pool.Exec(ctx, initSchema); err != nil {
		pool.Close()
		t.Fatalf("failed to apply schema: %v", err)
	}
	pool.Close()

	repo, err := repository.NewUserPostgresRepository(connStr)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	return repo
}

// newUser returns a valid domain.User with a unique username, phone, and email.
func newUser(suffix string) *domain.User {
	return &domain.User{
		Id:          uuid.New(),
		Username:    "testuser_" + suffix,
		PhoneNumber: "+20" + suffix, // Ensure it's not empty/colliding
		Email:       suffix + "@example.com",
		DisplayName: "Test User " + suffix,
		Bio:         "bio",
		BioStatus:   "active",
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestCreate_HappyPath(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u := newUser("create_ok")
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: unexpected error: %v", err)
	}
}

func TestCreate_NilUser_ReturnsInvalidInput(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	err := repo.Create(ctx, nil)
	if err != domain.ErrInvalidUserInput {
		t.Fatalf("want ErrInvalidUserInput, got %v", err)
	}
}

func TestCreate_DuplicateUsername_ReturnsAlreadyExists(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u := newUser("dup")
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("first Create: %v", err)
	}

	// same username, new id
	u2 := newUser("dup")
	u2.Id = uuid.New()

	err := repo.Create(ctx, u2)
	if err == nil {
		t.Fatal("expected error for duplicate username, got nil")
	}
}

// ── Get ───────────────────────────────────────────────────────────────────────

func TestGet_ByID_Found(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u := newUser("get_by_id")
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.Get(ctx, domain.Lookup{Type: domain.LookUpId, Value: u.Id.String()})
	if err != nil {
		t.Fatalf("Get: unexpected error: %v", err)
	}
	if got.Id != u.Id {
		t.Fatalf("want id %v, got %v", u.Id, got.Id)
	}
}

func TestGet_ByUsername_Found(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u := newUser("get_by_username")
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.Get(ctx, domain.Lookup{Type: domain.LookupUsername, Value: u.Username})
	if err != nil {
		t.Fatalf("Get: unexpected error: %v", err)
	}
	if got.Username != u.Username {
		t.Fatalf("want username %q, got %q", u.Username, got.Username)
	}
}

func TestGet_NotFound_ReturnsErrNotFound(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	_, err := repo.Get(ctx, domain.Lookup{Type: domain.LookUpId, Value: uuid.NewString()})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ── GetUsers ──────────────────────────────────────────────────────────────────

func TestGetUsers_EmptyIDs_ReturnsInvalidInput(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	_, err := repo.GetUsers(ctx, []string{})
	if err != utils_postgres.ErrInvalidInput {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}

func TestGetUsers_ReturnsMatchingUsers(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u1 := newUser("getusers_1")
	u2 := newUser("getusers_2")
	for _, u := range []*domain.User{u1, u2} {
		if err := repo.Create(ctx, u); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	users, err := repo.GetUsers(ctx, []string{u1.Id.String(), u2.Id.String()})
	if err != nil {
		t.Fatalf("GetUsers: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("want 2 users, got %d", len(users))
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestUpdate_HappyPath(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u := newUser("update_ok")
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	newBio := "updated bio"
	err := repo.Update(ctx, &domain.UserUpdate{
		Id:  u.Id.String(),
		Bio: &newBio,
	})
	if err != nil {
		t.Fatalf("Update: unexpected error: %v", err)
	}

	// verify the update persisted
	got, err := repo.Get(ctx, domain.Lookup{Type: domain.LookUpId, Value: u.Id.String()})
	if err != nil {
		t.Fatalf("Get after Update: %v", err)
	}
	if got.Bio != newBio {
		t.Fatalf("want bio %q, got %q", newBio, got.Bio)
	}
}

func TestUpdate_NoFields_ReturnsInvalidInput(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	err := repo.Update(ctx, &domain.UserUpdate{Id: uuid.NewString()})
	if err != utils_postgres.ErrInvalidInput {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}

// ── SoftDelete & RestoreUser ──────────────────────────────────────────────────

func TestSoftDelete_HappyPath(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u := newUser("soft_delete")
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.SoftDelete(ctx, u.Id.String()); err != nil {
		t.Fatalf("SoftDelete: unexpected error: %v", err)
	}
}

func TestSoftDelete_EmptyID_ReturnsInvalidInput(t *testing.T) {
	repo := newTestRepo(t)
	err := repo.SoftDelete(context.Background(), "")
	if err != utils_postgres.ErrInvalidInput {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}

func TestSoftDelete_UnknownID_ReturnsNotFound(t *testing.T) {
	repo := newTestRepo(t)
	err := repo.SoftDelete(context.Background(), uuid.NewString())
	if err != utils_postgres.ErrNotFound {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestRestoreUser_HappyPath(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u := newUser("restore")
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.SoftDelete(ctx, u.Id.String()); err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}
	if err := repo.RestoreUser(ctx, u.Id.String()); err != nil {
		t.Fatalf("RestoreUser: unexpected error: %v", err)
	}
}

func TestRestoreUser_EmptyID_ReturnsInvalidInput(t *testing.T) {
	repo := newTestRepo(t)
	err := repo.RestoreUser(context.Background(), "")
	if err != utils_postgres.ErrInvalidInput {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}

// ── GetSummary ────────────────────────────────────────────────────────────────

func TestGetSummary_Found(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u := newUser("summary_ok")
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	summary, err := repo.GetSummary(ctx, u.Id.String())
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	if summary.Id != u.Id.String() {
		t.Fatalf("want id %v, got %v", u.Id, summary.Id)
	}
}

func TestGetSummary_NotFound_ReturnsErrUserNotFound(t *testing.T) {
	repo := newTestRepo(t)
	_, err := repo.GetSummary(context.Background(), uuid.NewString())
	if err != domain.ErrUserNotFound {
		t.Fatalf("want ErrUserNotFound, got %v", err)
	}
}

func TestGetSummary_EmptyID_ReturnsInvalidInput(t *testing.T) {
	repo := newTestRepo(t)
	_, err := repo.GetSummary(context.Background(), "")
	if err != utils_postgres.ErrInvalidInput {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}

// ── GetSummaries ──────────────────────────────────────────────────────────────

func TestGetSummaries_EmptyIDs_ReturnsNil(t *testing.T) {
	repo := newTestRepo(t)
	result, err := repo.GetSummaries(context.Background(), []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("want nil, got %v", result)
	}
}

func TestGetSummaries_ReturnsSummaries(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u1 := newUser("summaries_1")
	u2 := newUser("summaries_2")
	for _, u := range []*domain.User{u1, u2} {
		if err := repo.Create(ctx, u); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	results, err := repo.GetSummaries(ctx, []string{u1.Id.String(), u2.Id.String()})
	if err != nil {
		t.Fatalf("GetSummaries: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("want 2 summaries, got %d", len(results))
	}
}

// ── Search ────────────────────────────────────────────────────────────────────

func TestSearch_MatchesByUsername(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u := newUser("searchable")
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	result, err := repo.Search(ctx, &domain.UserSearch{
		Query:    "searchable",
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(result.Users) == 0 {
		t.Fatal("want at least 1 result, got 0")
	}
}

func TestSearch_Pagination_HasMore(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	// Insert 3 users all matching "paginate"
	for i := 0; i < 3; i++ {
		u := newUser(fmt.Sprintf("paginate_%d", i))
		if err := repo.Create(ctx, u); err != nil {
			t.Fatalf("Create user %d: %v", i, err)
		}
	}

	result, err := repo.Search(ctx, &domain.UserSearch{
		Query:    "paginate",
		Page:     1,
		PageSize: 2,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if !result.HasMore {
		t.Fatal("want HasMore=true when there are more results than page size")
	}
	if len(result.Users) != 2 {
		t.Fatalf("want 2 users on page 1, got %d", len(result.Users))
	}
}

// ── Exists ────────────────────────────────────────────────────────────────────

func TestExists_Found(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u := newUser("exists_ok")
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	id, err := repo.Exists(ctx, domain.Lookup{Type: domain.LookUpId, Value: u.Id.String()})
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if id != u.Id.String() {
		t.Fatalf("want id %q, got %q", u.Id.String(), id)
	}
}

func TestExists_NotFound_ReturnsErrNotFound(t *testing.T) {
	repo := newTestRepo(t)
	_, err := repo.Exists(context.Background(), domain.Lookup{Type: domain.LookUpId, Value: uuid.NewString()})
	if err != utils_postgres.ErrNotFound {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

// ── UsersExist ────────────────────────────────────────────────────────────────

func TestUsersExist_EmptyIDs_ReturnsNil(t *testing.T) {
	repo := newTestRepo(t)
	result, err := repo.UsersExist(context.Background(), []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("want nil, got %v", result)
	}
}

func TestUsersExist_PartialMatch(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u := newUser("usersexist")
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	ghost := uuid.NewString()
	results, err := repo.UsersExist(ctx, []string{u.Id.String(), ghost})
	if err != nil {
		t.Fatalf("UsersExist: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("want 2 existence records, got %d", len(results))
	}

	found := map[string]bool{}
	for _, r := range results {
		found[r.Id] = r.Exists
	}
	if !found[u.Id.String()] {
		t.Errorf("want existing user %s to be found", u.Id)
	}
	if found[ghost] {
		t.Errorf("want ghost user %s to not be found", ghost)
	}
}

// ── RegisterContact ───────────────────────────────────────────────────────────

func TestRegisterContact_HappyPath_Email(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u := newUser("contact_email")
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	err := repo.RegisterContact(ctx, &domain.ContactRequest{
		UserId: u.Id.String(),
		ContactLookup: domain.ContactLookup{
			Type:  domain.ContactLookupTypeEmail,
			Value: "test@example.com",
		},
	})
	if err != nil {
		t.Fatalf("RegisterContact: unexpected error: %v", err)
	}
}

func TestRegisterContact_HappyPath_Phone(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u := newUser("contact_phone")
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	err := repo.RegisterContact(ctx, &domain.ContactRequest{
		UserId: u.Id.String(),
		ContactLookup: domain.ContactLookup{
			Type:  domain.ContactLookupTypePhone,
			Value: "+201234567890",
		},
	})
	if err != nil {
		t.Fatalf("RegisterContact: unexpected error: %v", err)
	}
}

func TestRegisterContact_EmptyUserID_ReturnsInvalidInput(t *testing.T) {
	repo := newTestRepo(t)
	err := repo.RegisterContact(context.Background(), &domain.ContactRequest{
		UserId: "",
		ContactLookup: domain.ContactLookup{
			Type:  domain.ContactLookupTypeEmail,
			Value: "x@x.com",
		},
	})
	if err != utils_postgres.ErrInvalidInput {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}

func TestRegisterContact_UnknownUserID_ReturnsNotFound(t *testing.T) {
	repo := newTestRepo(t)
	err := repo.RegisterContact(context.Background(), &domain.ContactRequest{
		UserId: uuid.NewString(),
		ContactLookup: domain.ContactLookup{
			Type:  domain.ContactLookupTypePhone,
			Value: "+201234567890",
		},
	})
	if err != utils_postgres.ErrNotFound {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

// ── Privacy Settings ──────────────────────────────────────────────────────────

func TestGetPrivacySettings_DefaultsAfterCreate(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u := newUser("privacy_get")
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	settings, err := repo.GetPrivacySettings(ctx, u.Id.String())
	if err != nil {
		t.Fatalf("GetPrivacySettings: %v", err)
	}
	// defaults from schema
	if settings.AvatarVisibility != domain.EVERYONE {
		t.Errorf("want AvatarVisibility=everyone, got %v", settings.AvatarVisibility)
	}
	if !settings.FindByUsername {
		t.Error("want FindByUsername=true by default")
	}
}

func TestUpdatePrivacySettings_HappyPath(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	u := newUser("privacy_update")
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create: %v", err)
	}

	nobody := domain.NOBODY
	err := repo.UpdatePrivacySettings(ctx, &domain.PrivacySettingsUpdate{
		UserId:           u.Id.String(),
		AvatarVisibility: &nobody,
	})
	if err != nil {
		t.Fatalf("UpdatePrivacySettings: %v", err)
	}

	settings, err := repo.GetPrivacySettings(ctx, u.Id.String())
	if err != nil {
		t.Fatalf("GetPrivacySettings after update: %v", err)
	}
	if settings.AvatarVisibility != domain.NOBODY {
		t.Errorf("want AvatarVisibility=nobody after update, got %v", settings.AvatarVisibility)
	}
}

func TestUpdatePrivacySettings_NoFields_ReturnsInvalidInput(t *testing.T) {
	repo := newTestRepo(t)
	err := repo.UpdatePrivacySettings(context.Background(), &domain.PrivacySettingsUpdate{
		UserId: uuid.NewString(),
	})
	if err != utils_postgres.ErrInvalidInput {
		t.Fatalf("want ErrInvalidInput, got %v", err)
	}
}
