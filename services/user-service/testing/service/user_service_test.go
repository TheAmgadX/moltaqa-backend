package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/domain"
	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/infrastructure/repository"
	servicepkg "github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/service"
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

// ── fakeRepo ────────────────────────────────────────────────────────────────

type fakeRepo struct {
	createdUsers     []*domain.User
	contactRequests  []*domain.ContactRequest
	updatedUsers     []*domain.UserUpdate
	deletedIDs       []string
	restoredIDs      []string
	getUser          *domain.User
	getUsers         []domain.User
	getSummary       *domain.UserSummary
	getSummaries     []domain.UserSummary
	searchResult     *domain.UserSearchResult
	existsID         string
	usersExistResult []domain.UserExistence
	privacySettings  *domain.PrivacySettings

	// injectable errors per operation
	createErr        error
	contactErr       error
	updateErr        error
	softDeleteErr    error
	restoreErr       error
	getErr           error
	getUsersErr      error
	getSummaryErr    error
	getSummariesErr  error
	searchErr        error
	existsErr        error
	usersExistErr    error
	privacyErr       error
	updatePrivacyErr error
}

func (r *fakeRepo) Create(_ context.Context, user *domain.User) error {
	r.createdUsers = append(r.createdUsers, user)
	return r.createErr
}
func (r *fakeRepo) RegisterContact(_ context.Context, contact *domain.ContactRequest) error {
	r.contactRequests = append(r.contactRequests, contact)
	return r.contactErr
}
func (r *fakeRepo) Update(_ context.Context, update *domain.UserUpdate) error {
	r.updatedUsers = append(r.updatedUsers, update)
	return r.updateErr
}
func (r *fakeRepo) SoftDelete(_ context.Context, id string) error {
	r.deletedIDs = append(r.deletedIDs, id)
	return r.softDeleteErr
}
func (r *fakeRepo) RestoreUser(_ context.Context, id string) error {
	r.restoredIDs = append(r.restoredIDs, id)
	return r.restoreErr
}
func (r *fakeRepo) Get(_ context.Context, _ domain.Lookup) (*domain.User, error) {
	return r.getUser, r.getErr
}
func (r *fakeRepo) GetUsers(_ context.Context, _ []string) ([]domain.User, error) {
	return r.getUsers, r.getUsersErr
}
func (r *fakeRepo) GetSummary(_ context.Context, _ string) (*domain.UserSummary, error) {
	return r.getSummary, r.getSummaryErr
}
func (r *fakeRepo) GetSummaries(_ context.Context, _ []string) ([]domain.UserSummary, error) {
	return r.getSummaries, r.getSummariesErr
}
func (r *fakeRepo) Search(_ context.Context, _ *domain.UserSearch) (*domain.UserSearchResult, error) {
	return r.searchResult, r.searchErr
}
func (r *fakeRepo) Exists(_ context.Context, _ domain.Lookup) (string, error) {
	return r.existsID, r.existsErr
}
func (r *fakeRepo) UsersExist(_ context.Context, _ []string) ([]domain.UserExistence, error) {
	return r.usersExistResult, r.usersExistErr
}
func (r *fakeRepo) GetPrivacySettings(_ context.Context, _ string) (*domain.PrivacySettings, error) {
	return r.privacySettings, r.privacyErr
}
func (r *fakeRepo) UpdatePrivacySettings(_ context.Context, _ *domain.PrivacySettingsUpdate) error {
	return r.updatePrivacyErr
}

// compile-time interface guard
var _ repository.UserRepository = (*fakeRepo)(nil)

// ── helpers ─────────────────────────────────────────────────────────────────

// mustNewService constructs a UserService backed by fakeRepo and a real Kafka
// client pointed at the local broker started by Docker Compose.
func mustNewService(t *testing.T, repo repository.UserRepository) *servicepkg.UserService {
	t.Helper()
	client, err := kgo.NewClient(
		kgo.SeedBrokers("127.0.0.1:9092"),
		kgo.RecordDeliveryTimeout(1*time.Second),
		kgo.RequestTimeoutOverhead(100*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("failed to construct kafka client for tests: %v", err)
	}
	t.Cleanup(client.Close)

	svc, err := servicepkg.NewService(repo, client)
	if err != nil {
		t.Fatalf("failed to create service under test: %v", err)
	}
	return svc
}

// testContext returns a context with a short but generous timeout appropriate
// for unit tests (no real DB round-trip is involved).
func testContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}

// ptr returns a pointer to any value — reduces noise in test fixtures.
func ptr[T any](v T) *T { return &v }

// testUser returns a minimal valid User suitable as a create payload.
func testUser() *domain.User {
	return &domain.User{Email: "user@example.com", PhoneNumber: "+12065550199"}
}

// testUserUpdate returns a fully-populated UserUpdate for a given id.
func testUserUpdate(id string) *domain.UserUpdate {
	return &domain.UserUpdate{
		Id:              id,
		Bio:             ptr("test bio"),
		BioStatus:       ptr("online"),
		DisplayName:     ptr("Updated User"),
		Username:        ptr("updated_user"),
		ProfileImageUrl: ptr("https://cdn.example.com/avatar.png"),
		BirthDate:       ptr(time.Now().Add(-24 * time.Hour)),
	}
}

// ── Create ───────────────────────────────────────────────────────────────────

func TestUserService_Create(t *testing.T) {
	t.Run("happy path — email only", func(t *testing.T) {
		repo := &fakeRepo{}
		svc := mustNewService(t, repo)
		user := &domain.User{Email: "user@example.com"}
		ctx, cancel := testContext()
		defer cancel()

		if err := svc.Create(ctx, user); err != nil {
			t.Fatalf("Create() unexpected error: %v", err)
		}
		if len(repo.createdUsers) != 1 {
			t.Fatalf("expected repo.Create called once, got %d", len(repo.createdUsers))
		}
		if user.Id == uuid.Nil {
			t.Fatal("expected service to assign a UUID on create")
		}
	})

	t.Run("happy path — phone only", func(t *testing.T) {
		repo := &fakeRepo{}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		if err := svc.Create(ctx, &domain.User{PhoneNumber: "+12065550199"}); err != nil {
			t.Fatalf("Create() phone-only unexpected error: %v", err)
		}
	})

	t.Run("happy path — email and phone", func(t *testing.T) {
		repo := &fakeRepo{}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		if err := svc.Create(ctx, testUser()); err != nil {
			t.Fatalf("Create() email+phone unexpected error: %v", err)
		}
	})

	t.Run("error — neither email nor phone", func(t *testing.T) {
		svc := mustNewService(t, &fakeRepo{})
		ctx, cancel := testContext()
		defer cancel()

		if err := svc.Create(ctx, &domain.User{}); err == nil {
			t.Fatal("expected Create() to reject a user with no credentials")
		}
	})

	t.Run("error — invalid email format", func(t *testing.T) {
		svc := mustNewService(t, &fakeRepo{})
		ctx, cancel := testContext()
		defer cancel()

		if err := svc.Create(ctx, &domain.User{Email: "not-an-email"}); err == nil {
			t.Fatal("expected Create() to reject an invalid email")
		}
	})

	t.Run("error — invalid phone format", func(t *testing.T) {
		svc := mustNewService(t, &fakeRepo{})
		ctx, cancel := testContext()
		defer cancel()

		if err := svc.Create(ctx, &domain.User{PhoneNumber: "12345"}); err == nil {
			t.Fatal("expected Create() to reject an invalid phone number")
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		repoErr := errors.New("db down")
		repo := &fakeRepo{createErr: repoErr}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		if err := svc.Create(ctx, testUser()); err == nil {
			t.Fatal("expected Create() to propagate repo error")
		}
	})
}

// ── RegisterContact ──────────────────────────────────────────────────────────

func TestUserService_RegisterContact(t *testing.T) {
	t.Run("happy path — email contact", func(t *testing.T) {
		repo := &fakeRepo{}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		contact := &domain.ContactRequest{
			UserId: "user-123",
			ContactLookup: domain.ContactLookup{
				Type:  domain.ContactLookupTypeEmail,
				Value: "contact@example.com",
			},
		}
		if err := svc.RegisterContact(ctx, contact); err != nil {
			t.Fatalf("RegisterContact() unexpected error: %v", err)
		}
		if len(repo.contactRequests) != 1 {
			t.Fatalf("expected repo.RegisterContact once, got %d", len(repo.contactRequests))
		}
	})

	t.Run("happy path — phone contact", func(t *testing.T) {
		repo := &fakeRepo{}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		contact := &domain.ContactRequest{
			UserId: "user-123",
			ContactLookup: domain.ContactLookup{
				Type:  domain.ContactLookupTypePhone,
				Value: "+12065550199",
			},
		}
		if err := svc.RegisterContact(ctx, contact); err != nil {
			t.Fatalf("RegisterContact() phone unexpected error: %v", err)
		}
	})

	t.Run("error — empty contact value", func(t *testing.T) {
		svc := mustNewService(t, &fakeRepo{})
		ctx, cancel := testContext()
		defer cancel()

		contact := &domain.ContactRequest{UserId: "user-123"}
		if err := svc.RegisterContact(ctx, contact); err == nil {
			t.Fatal("expected RegisterContact() to reject an empty contact value")
		}
	})

	t.Run("error — invalid email contact", func(t *testing.T) {
		svc := mustNewService(t, &fakeRepo{})
		ctx, cancel := testContext()
		defer cancel()

		contact := &domain.ContactRequest{
			UserId: "user-123",
			ContactLookup: domain.ContactLookup{
				Type:  domain.ContactLookupTypeEmail,
				Value: "not-an-email",
			},
		}
		if err := svc.RegisterContact(ctx, contact); err == nil {
			t.Fatal("expected RegisterContact() to reject an invalid email")
		}
	})

	t.Run("error — invalid phone contact", func(t *testing.T) {
		svc := mustNewService(t, &fakeRepo{})
		ctx, cancel := testContext()
		defer cancel()

		contact := &domain.ContactRequest{
			UserId: "user-123",
			ContactLookup: domain.ContactLookup{
				Type:  domain.ContactLookupTypePhone,
				Value: "12345",
			},
		}
		if err := svc.RegisterContact(ctx, contact); err == nil {
			t.Fatal("expected RegisterContact() to reject an invalid phone")
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		repo := &fakeRepo{contactErr: errors.New("db error")}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		contact := &domain.ContactRequest{
			UserId: "user-123",
			ContactLookup: domain.ContactLookup{
				Type:  domain.ContactLookupTypeEmail,
				Value: "contact@example.com",
			},
		}
		if err := svc.RegisterContact(ctx, contact); err == nil {
			t.Fatal("expected RegisterContact() to propagate repo error")
		}
	})
}

// ── Update ───────────────────────────────────────────────────────────────────

func TestUserService_Update(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo := &fakeRepo{}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		update := testUserUpdate(uuid.NewString())
		if err := svc.Update(ctx, update); err != nil {
			t.Fatalf("Update() unexpected error: %v", err)
		}
		if len(repo.updatedUsers) != 1 {
			t.Fatalf("expected repo.Update once, got %d", len(repo.updatedUsers))
		}
	})

	t.Run("error — invalid UUID", func(t *testing.T) {
		svc := mustNewService(t, &fakeRepo{})
		ctx, cancel := testContext()
		defer cancel()

		update := &domain.UserUpdate{Id: "not-a-uuid"}
		if err := svc.Update(ctx, update); err == nil {
			t.Fatal("expected Update() to reject an invalid user ID")
		}
	})

	t.Run("error — future birth date", func(t *testing.T) {
		svc := mustNewService(t, &fakeRepo{})
		ctx, cancel := testContext()
		defer cancel()

		update := &domain.UserUpdate{Id: uuid.NewString(), BirthDate: ptr(time.Now().Add(24 * time.Hour))}
		if err := svc.Update(ctx, update); err == nil {
			t.Fatal("expected Update() to reject a future birth date")
		}
	})

	t.Run("error — bio too long", func(t *testing.T) {
		svc := mustNewService(t, &fakeRepo{})
		ctx, cancel := testContext()
		defer cancel()

		longBio := string(make([]byte, 251))
		update := &domain.UserUpdate{Id: uuid.NewString(), Bio: &longBio}
		if err := svc.Update(ctx, update); err == nil {
			t.Fatal("expected Update() to reject a bio exceeding 250 characters")
		}
	})

	t.Run("error — username with special characters", func(t *testing.T) {
		svc := mustNewService(t, &fakeRepo{})
		ctx, cancel := testContext()
		defer cancel()

		update := &domain.UserUpdate{Id: uuid.NewString(), Username: ptr("invalid user name!")}
		if err := svc.Update(ctx, update); err == nil {
			t.Fatal("expected Update() to reject a username with special characters")
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		repo := &fakeRepo{updateErr: errors.New("db error")}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		if err := svc.Update(ctx, testUserUpdate(uuid.NewString())); err == nil {
			t.Fatal("expected Update() to propagate repo error")
		}
	})
}

// ── Delete ───────────────────────────────────────────────────────────────────

func TestUserService_Delete(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo := &fakeRepo{}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		if err := svc.Delete(ctx, "user-123"); err != nil {
			t.Fatalf("Delete() unexpected error: %v", err)
		}
		if len(repo.deletedIDs) != 1 || repo.deletedIDs[0] != "user-123" {
			t.Fatalf("expected SoftDelete for user-123, got %#v", repo.deletedIDs)
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		repo := &fakeRepo{softDeleteErr: errors.New("db error")}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		if err := svc.Delete(ctx, "user-123"); err == nil {
			t.Fatal("expected Delete() to propagate repo error")
		}
	})
}

// ── Restore ──────────────────────────────────────────────────────────────────

func TestUserService_Restore(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo := &fakeRepo{}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		if err := svc.Restore(ctx, "user-123"); err != nil {
			t.Fatalf("Restore() unexpected error: %v", err)
		}
		if len(repo.restoredIDs) != 1 || repo.restoredIDs[0] != "user-123" {
			t.Fatalf("expected RestoreUser for user-123, got %#v", repo.restoredIDs)
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		repo := &fakeRepo{restoreErr: errors.New("db error")}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		if err := svc.Restore(ctx, "user-123"); err == nil {
			t.Fatal("expected Restore() to propagate repo error")
		}
	})
}

// ── Get ──────────────────────────────────────────────────────────────────────

func TestUserService_Get(t *testing.T) {
	t.Run("happy path — lookup by email", func(t *testing.T) {
		repo := &fakeRepo{getUser: &domain.User{Email: "lookup@example.com"}}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		got, _, err := svc.Get(ctx, domain.Lookup{Type: domain.LookupEmail, Value: "lookup@example.com"})
		if err != nil {
			t.Fatalf("Get() unexpected error: %v", err)
		}
		if got == nil || got.Email != "lookup@example.com" {
			t.Fatal("expected Get() to return the requested user")
		}
	})

	t.Run("happy path — lookup by ID", func(t *testing.T) {
		repo := &fakeRepo{getUser: &domain.User{Email: "id-user@example.com"}}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		got, _, err := svc.Get(ctx, domain.Lookup{Type: domain.LookUpId, Value: "some-id"})
		if err != nil {
			t.Fatalf("Get() by id unexpected error: %v", err)
		}
		if got == nil {
			t.Fatal("expected Get() by ID to return a user")
		}
	})

	t.Run("error — invalid email lookup", func(t *testing.T) {
		svc := mustNewService(t, &fakeRepo{})
		ctx, cancel := testContext()
		defer cancel()

		_, _, err := svc.Get(ctx, domain.Lookup{Type: domain.LookupEmail, Value: "not-an-email"})
		if err == nil {
			t.Fatal("expected Get() to reject an invalid email lookup")
		}
	})

	t.Run("error — repo not found propagated", func(t *testing.T) {
		repo := &fakeRepo{getErr: errors.New("not found")}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		_, _, err := svc.Get(ctx, domain.Lookup{Type: domain.LookupEmail, Value: "a@b.com"})
		if err == nil {
			t.Fatal("expected Get() to propagate repo error")
		}
	})
}

// ── GetUsers ─────────────────────────────────────────────────────────────────

func TestUserService_GetUsers(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo := &fakeRepo{getUsers: []domain.User{{Email: "one@example.com"}, {Email: "two@example.com"}}}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		got, err := svc.GetUsers(ctx, []string{"1", "2"})
		if err != nil {
			t.Fatalf("GetUsers() unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("expected two users, got %d", len(got))
		}
	})

	t.Run("empty ids returns nil without calling repo", func(t *testing.T) {
		repo := &fakeRepo{}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		got, err := svc.GetUsers(ctx, []string{})
		if err != nil {
			t.Fatalf("GetUsers() empty unexpected error: %v", err)
		}
		if got != nil {
			t.Fatal("expected nil result for empty IDs")
		}
		if len(repo.createdUsers) != 0 {
			t.Fatal("expected no repo calls for empty ids")
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		repo := &fakeRepo{getUsersErr: errors.New("db error")}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		_, err := svc.GetUsers(ctx, []string{"1", "2"})
		if err == nil {
			t.Fatal("expected GetUsers() to propagate repo error")
		}
	})
}

// ── GetUserSummary ───────────────────────────────────────────────────────────

func TestUserService_GetUserSummary(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo := &fakeRepo{getSummary: &domain.UserSummary{Id: "user-123", Username: "summary_user"}}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		got, err := svc.GetUserSummary(ctx, "user-123")
		if err != nil {
			t.Fatalf("GetUserSummary() unexpected error: %v", err)
		}
		if got == nil || got.Id != "user-123" {
			t.Fatal("expected GetUserSummary() to return the summary object")
		}
	})

	t.Run("error — empty ID", func(t *testing.T) {
		svc := mustNewService(t, &fakeRepo{})
		ctx, cancel := testContext()
		defer cancel()

		_, err := svc.GetUserSummary(ctx, "")
		if err == nil {
			t.Fatal("expected GetUserSummary() to reject an empty ID")
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		repo := &fakeRepo{getSummaryErr: errors.New("db error")}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		_, err := svc.GetUserSummary(ctx, "user-123")
		if err == nil {
			t.Fatal("expected GetUserSummary() to propagate repo error")
		}
	})
}

// ── GetUsersSummary ──────────────────────────────────────────────────────────

func TestUserService_GetUsersSummary(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo := &fakeRepo{getSummaries: []domain.UserSummary{{Id: "1"}, {Id: "2"}}}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		got, err := svc.GetUsersSummary(ctx, []string{"1", "2"})
		if err != nil {
			t.Fatalf("GetUsersSummary() unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("expected two summaries, got %d", len(got))
		}
	})

	t.Run("empty ids returns nil without calling repo", func(t *testing.T) {
		repo := &fakeRepo{}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		got, err := svc.GetUsersSummary(ctx, []string{})
		if err != nil || got != nil {
			t.Fatalf("expected (nil, nil) for empty ids, got (%v, %v)", got, err)
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		repo := &fakeRepo{getSummariesErr: errors.New("db error")}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		_, err := svc.GetUsersSummary(ctx, []string{"1", "2"})
		if err == nil {
			t.Fatal("expected GetUsersSummary() to propagate repo error")
		}
	})
}

// ── SearchUsers ──────────────────────────────────────────────────────────────

func TestUserService_SearchUsers(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo := &fakeRepo{searchResult: &domain.UserSearchResult{HasMore: true, Users: []domain.UserSummary{{Id: "search-01"}}}}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		got, err := svc.SearchUsers(ctx, &domain.UserSearch{Query: "alice", Page: 1, PageSize: 10})
		if err != nil {
			t.Fatalf("SearchUsers() unexpected error: %v", err)
		}
		if got == nil || !got.HasMore || len(got.Users) != 1 {
			t.Fatal("expected SearchUsers() to return a populated result")
		}
	})

	t.Run("nil request returns nil without calling repo", func(t *testing.T) {
		repo := &fakeRepo{}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		got, err := svc.SearchUsers(ctx, nil)
		if err != nil || got != nil {
			t.Fatalf("expected (nil, nil) for nil request, got (%v, %v)", got, err)
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		repo := &fakeRepo{searchErr: errors.New("db error")}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		_, err := svc.SearchUsers(ctx, &domain.UserSearch{Query: "alice"})
		if err == nil {
			t.Fatal("expected SearchUsers() to propagate repo error")
		}
	})
}

// ── UserExists ───────────────────────────────────────────────────────────────

func TestUserService_UserExists(t *testing.T) {
	t.Run("happy path — user exists", func(t *testing.T) {
		repo := &fakeRepo{existsID: "exists-user"}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		got, err := svc.UserExists(ctx, domain.Lookup{Type: domain.LookUpId, Value: "exists-user"})
		if err != nil {
			t.Fatalf("UserExists() unexpected error: %v", err)
		}
		if got.Id != "exists-user" || !got.Exists {
			t.Fatal("expected UserExists() to return a positive existence result")
		}
	})

	t.Run("error — invalid email lookup", func(t *testing.T) {
		svc := mustNewService(t, &fakeRepo{})
		ctx, cancel := testContext()
		defer cancel()

		_, err := svc.UserExists(ctx, domain.Lookup{Type: domain.LookupEmail, Value: "not-an-email"})
		if err == nil {
			t.Fatal("expected UserExists() to reject an invalid email lookup")
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		repo := &fakeRepo{existsErr: errors.New("db error")}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		_, err := svc.UserExists(ctx, domain.Lookup{Type: domain.LookUpId, Value: "user-123"})
		if err == nil {
			t.Fatal("expected UserExists() to propagate repo error")
		}
	})
}

// ── UsersExist ───────────────────────────────────────────────────────────────

func TestUserService_UsersExist(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo := &fakeRepo{usersExistResult: []domain.UserExistence{{Id: "a", Exists: true}, {Id: "b", Exists: false}}}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		got, err := svc.UsersExist(ctx, []string{"a", "b"})
		if err != nil {
			t.Fatalf("UsersExist() unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("expected two existence results, got %d", len(got))
		}
	})

	t.Run("empty ids returns nil without calling repo", func(t *testing.T) {
		svc := mustNewService(t, &fakeRepo{})
		ctx, cancel := testContext()
		defer cancel()

		got, err := svc.UsersExist(ctx, []string{})
		if err != nil || got != nil {
			t.Fatalf("expected (nil, nil) for empty ids, got (%v, %v)", got, err)
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		repo := &fakeRepo{usersExistErr: errors.New("db error")}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		_, err := svc.UsersExist(ctx, []string{"a", "b"})
		if err == nil {
			t.Fatal("expected UsersExist() to propagate repo error")
		}
	})
}

// ── GetPrivacySettings ───────────────────────────────────────────────────────

func TestUserService_GetPrivacySettings(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo := &fakeRepo{privacySettings: &domain.PrivacySettings{UserId: uuid.New(), AvatarVisibility: domain.EVERYONE}}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		got, err := svc.GetPrivacySettings(ctx, "user-123")
		if err != nil {
			t.Fatalf("GetPrivacySettings() unexpected error: %v", err)
		}
		if got == nil || got.AvatarVisibility != domain.EVERYONE {
			t.Fatal("expected GetPrivacySettings() to return a populated settings object")
		}
	})

	t.Run("error — empty user ID", func(t *testing.T) {
		svc := mustNewService(t, &fakeRepo{})
		ctx, cancel := testContext()
		defer cancel()

		_, err := svc.GetPrivacySettings(ctx, "")
		if err == nil {
			t.Fatal("expected GetPrivacySettings() to reject an empty user ID")
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		repo := &fakeRepo{privacyErr: errors.New("db error")}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		_, err := svc.GetPrivacySettings(ctx, "user-123")
		if err == nil {
			t.Fatal("expected GetPrivacySettings() to propagate repo error")
		}
	})
}

// ── UpdatePrivacySettings ────────────────────────────────────────────────────

func TestUserService_UpdatePrivacySettings(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo := &fakeRepo{}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		if err := svc.UpdatePrivacySettings(ctx, &domain.PrivacySettingsUpdate{
			UserId:           "user-123",
			AvatarVisibility: ptr(domain.EVERYONE),
		}); err != nil {
			t.Fatalf("UpdatePrivacySettings() unexpected error: %v", err)
		}
	})

	t.Run("error — nil settings", func(t *testing.T) {
		svc := mustNewService(t, &fakeRepo{})
		ctx, cancel := testContext()
		defer cancel()

		if err := svc.UpdatePrivacySettings(ctx, nil); err == nil {
			t.Fatal("expected UpdatePrivacySettings() to reject nil settings")
		}
	})

	t.Run("error — invalid avatar visibility", func(t *testing.T) {
		svc := mustNewService(t, &fakeRepo{})
		ctx, cancel := testContext()
		defer cancel()

		invalid := domain.Visibility("invalid")
		if err := svc.UpdatePrivacySettings(ctx, &domain.PrivacySettingsUpdate{
			UserId:           "user-123",
			AvatarVisibility: &invalid,
		}); err == nil {
			t.Fatal("expected UpdatePrivacySettings() to reject an invalid visibility value")
		}
	})

	t.Run("error — invalid email visibility", func(t *testing.T) {
		svc := mustNewService(t, &fakeRepo{})
		ctx, cancel := testContext()
		defer cancel()

		invalid := domain.Visibility("bad-value")
		if err := svc.UpdatePrivacySettings(ctx, &domain.PrivacySettingsUpdate{
			UserId:          "user-123",
			EmailVisibility: &invalid,
		}); err == nil {
			t.Fatal("expected UpdatePrivacySettings() to reject an invalid email visibility")
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		repo := &fakeRepo{updatePrivacyErr: errors.New("db error")}
		svc := mustNewService(t, repo)
		ctx, cancel := testContext()
		defer cancel()

		if err := svc.UpdatePrivacySettings(ctx, &domain.PrivacySettingsUpdate{
			UserId:           "user-123",
			AvatarVisibility: ptr(domain.FRIENDS),
		}); err == nil {
			t.Fatal("expected UpdatePrivacySettings() to propagate repo error")
		}
	})
}
