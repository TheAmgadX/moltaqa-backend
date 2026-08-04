package grpc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/domain"
	grpcpkg "github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/infrastructure/grpc"
	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/infrastructure/repository"
	servicepkg "github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/service"
	pb "github.com/TheAmgadX/moltaqa-backend/shared/proto/users"
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ── grpcFakeRepo ─────────────────────────────────────────────────────────────

type grpcFakeRepo struct {
	user         *domain.User
	users        []domain.User
	summary      *domain.UserSummary
	summaries    []domain.UserSummary
	searchResult *domain.UserSearchResult
	existsID     string
	existence    []domain.UserExistence
	privacy      *domain.PrivacySettings

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

func (r *grpcFakeRepo) Create(_ context.Context, _ *domain.User) error { return r.createErr }
func (r *grpcFakeRepo) RegisterContact(_ context.Context, _ *domain.ContactRequest) error {
	return r.contactErr
}
func (r *grpcFakeRepo) Update(_ context.Context, _ *domain.UserUpdate) error { return r.updateErr }
func (r *grpcFakeRepo) SoftDelete(_ context.Context, _ string) error         { return r.softDeleteErr }
func (r *grpcFakeRepo) RestoreUser(_ context.Context, _ string) error        { return r.restoreErr }
func (r *grpcFakeRepo) Get(_ context.Context, _ domain.Lookup) (*domain.User, error) {
	return r.user, r.getErr
}
func (r *grpcFakeRepo) GetUsers(_ context.Context, _ []string) ([]domain.User, error) {
	return r.users, r.getUsersErr
}
func (r *grpcFakeRepo) GetSummary(_ context.Context, _ string) (*domain.UserSummary, error) {
	return r.summary, r.getSummaryErr
}
func (r *grpcFakeRepo) GetSummaries(_ context.Context, _ []string) ([]domain.UserSummary, error) {
	return r.summaries, r.getSummariesErr
}
func (r *grpcFakeRepo) Search(_ context.Context, _ *domain.UserSearch) (*domain.UserSearchResult, error) {
	return r.searchResult, r.searchErr
}
func (r *grpcFakeRepo) Exists(_ context.Context, _ domain.Lookup) (string, error) {
	return r.existsID, r.existsErr
}
func (r *grpcFakeRepo) UsersExist(_ context.Context, _ []string) ([]domain.UserExistence, error) {
	return r.existence, r.usersExistErr
}
func (r *grpcFakeRepo) GetPrivacySettings(_ context.Context, _ string) (*domain.PrivacySettings, error) {
	return r.privacy, r.privacyErr
}
func (r *grpcFakeRepo) UpdatePrivacySettings(_ context.Context, _ *domain.PrivacySettingsUpdate) error {
	return r.updatePrivacyErr
}

// compile-time interface guard
var _ repository.UserRepository = (*grpcFakeRepo)(nil)

// ── helpers ──────────────────────────────────────────────────────────────────

func mustNewGRPCServer(t *testing.T, repo *grpcFakeRepo) *grpcpkg.UserGRPCServer {
	t.Helper()
	client, err := kgo.NewClient(
		kgo.SeedBrokers("127.0.0.1:9092"),
		kgo.RecordDeliveryTimeout(1*time.Second),
		kgo.RequestTimeoutOverhead(100*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("failed to construct kafka client for grpc tests: %v", err)
	}
	t.Cleanup(client.Close)

	svc, err := servicepkg.NewService(repo, client)
	if err != nil {
		t.Fatalf("failed to construct service for grpc tests: %v", err)
	}
	return grpcpkg.NewUserGRPCServer(svc)
}

func testCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}

func strPtr(v string) *string { return &v }

// ── CreateUser ───────────────────────────────────────────────────────────────

func TestUserGRPCServer_CreateUser(t *testing.T) {
	t.Run("happy path — email contact", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{})
		ctx, cancel := testCtx()
		defer cancel()

		resp, err := server.CreateUser(ctx, &pb.CreateUserRequest{
			Contact: &pb.CreateUserRequest_Email{Email: "user@example.com"},
		})
		if err != nil {
			t.Fatalf("CreateUser() unexpected error: %v", err)
		}
		if resp == nil || resp.User == nil || resp.User.Email != "user@example.com" {
			t.Fatal("expected CreateUser() to return the mapped user payload")
		}
	})

	t.Run("happy path — phone contact", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{})
		ctx, cancel := testCtx()
		defer cancel()

		resp, err := server.CreateUser(ctx, &pb.CreateUserRequest{
			Contact: &pb.CreateUserRequest_Phone{Phone: "+12065550199"},
		})
		if err != nil {
			t.Fatalf("CreateUser() phone unexpected error: %v", err)
		}
		if resp == nil || resp.User == nil {
			t.Fatal("expected CreateUser() phone to return a user payload")
		}
	})

	t.Run("error — invalid email", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.CreateUser(ctx, &pb.CreateUserRequest{
			Contact: &pb.CreateUserRequest_Email{Email: "not-an-email"},
		})
		if err == nil {
			t.Fatal("expected CreateUser() to reject an invalid email")
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{createErr: errors.New("db down")})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.CreateUser(ctx, &pb.CreateUserRequest{
			Contact: &pb.CreateUserRequest_Email{Email: "user@example.com"},
		})
		if err == nil {
			t.Fatal("expected CreateUser() to propagate repo error as gRPC status")
		}
	})
}

// ── RegisterContact ──────────────────────────────────────────────────────────

func TestUserGRPCServer_RegisterContact(t *testing.T) {
	t.Run("happy path — email", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.RegisterContact(ctx, &pb.RegisterContactRequest{
			UserId:      "user-123",
			ContactType: &pb.RegisterContactRequest_Email{Email: "contact@example.com"},
		})
		if err != nil {
			t.Fatalf("RegisterContact() unexpected error: %v", err)
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{contactErr: errors.New("db error")})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.RegisterContact(ctx, &pb.RegisterContactRequest{
			UserId:      "user-123",
			ContactType: &pb.RegisterContactRequest_Email{Email: "contact@example.com"},
		})
		if err == nil {
			t.Fatal("expected RegisterContact() to propagate repo error as gRPC status")
		}
	})
}

// ── UpdateUser ───────────────────────────────────────────────────────────────

func TestUserGRPCServer_UpdateUser(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{})
		ctx, cancel := testCtx()
		defer cancel()

		birthDate := timestamppb.New(time.Now().Add(-24 * time.Hour))
		_, err := server.UpdateUser(ctx, &pb.UpdateUserRequest{
			Id:          uuid.NewString(),
			Username:    strPtr("updated_user"),
			DisplayName: strPtr("Updated User"),
			Bio:         strPtr("bio"),
			BirthDate:   birthDate,
		})
		if err != nil {
			t.Fatalf("UpdateUser() unexpected error: %v", err)
		}
	})

	t.Run("error — invalid UUID", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.UpdateUser(ctx, &pb.UpdateUserRequest{
			Id: "not-a-uuid",
		})
		if err == nil {
			t.Fatal("expected UpdateUser() to reject an invalid UUID")
		}
	})

	t.Run("error — future birth date", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.UpdateUser(ctx, &pb.UpdateUserRequest{
			Id:        uuid.NewString(),
			BirthDate: timestamppb.New(time.Now().Add(24 * time.Hour)),
		})
		if err == nil {
			t.Fatal("expected UpdateUser() to reject a future birth date")
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{updateErr: errors.New("db error")})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.UpdateUser(ctx, &pb.UpdateUserRequest{
			Id:       uuid.NewString(),
			Username: strPtr("ok_user"),
		})
		if err == nil {
			t.Fatal("expected UpdateUser() to propagate repo error as gRPC status")
		}
	})
}

// ── DeleteUser ───────────────────────────────────────────────────────────────

func TestUserGRPCServer_DeleteUser(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.DeleteUser(ctx, &pb.DeleteUserRequest{Id: "user-123"})
		if err != nil {
			t.Fatalf("DeleteUser() unexpected error: %v", err)
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{softDeleteErr: errors.New("db error")})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.DeleteUser(ctx, &pb.DeleteUserRequest{Id: "user-123"})
		if err == nil {
			t.Fatal("expected DeleteUser() to propagate repo error as gRPC status")
		}
	})
}

// ── RestoreUser ──────────────────────────────────────────────────────────────

func TestUserGRPCServer_RestoreUser(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.RestoreUser(ctx, &pb.RestoreUserRequest{Id: "user-123"})
		if err != nil {
			t.Fatalf("RestoreUser() unexpected error: %v", err)
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{restoreErr: errors.New("db error")})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.RestoreUser(ctx, &pb.RestoreUserRequest{Id: "user-123"})
		if err == nil {
			t.Fatal("expected RestoreUser() to propagate repo error as gRPC status")
		}
	})
}

// ── GetUser ──────────────────────────────────────────────────────────────────

func TestUserGRPCServer_GetUser(t *testing.T) {
	t.Run("happy path — lookup by email", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{user: &domain.User{Email: "lookup@example.com"}})
		ctx, cancel := testCtx()
		defer cancel()

		resp, err := server.GetUser(ctx, &pb.GetUserRequest{
			Lookup: &pb.GetUserRequest_Email{Email: "lookup@example.com"},
		})
		if err != nil {
			t.Fatalf("GetUser() unexpected error: %v", err)
		}
		if resp == nil || resp.User == nil || resp.User.Email != "lookup@example.com" {
			t.Fatal("expected GetUser() to return the mapped user")
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{getErr: errors.New("not found")})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.GetUser(ctx, &pb.GetUserRequest{
			Lookup: &pb.GetUserRequest_Email{Email: "a@b.com"},
		})
		if err == nil {
			t.Fatal("expected GetUser() to propagate repo error as gRPC status")
		}
	})
}

// ── GetUsers ─────────────────────────────────────────────────────────────────

func TestUserGRPCServer_GetUsers(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{users: []domain.User{{Email: "one@example.com"}, {Email: "two@example.com"}}})
		ctx, cancel := testCtx()
		defer cancel()

		resp, err := server.GetUsers(ctx, &pb.GetUsersRequest{UserIds: []string{"1", "2"}})
		if err != nil {
			t.Fatalf("GetUsers() unexpected error: %v", err)
		}
		if len(resp.Users) != 2 {
			t.Fatalf("expected 2 users, got %d", len(resp.Users))
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{getUsersErr: errors.New("db error")})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.GetUsers(ctx, &pb.GetUsersRequest{UserIds: []string{"1", "2"}})
		if err == nil {
			t.Fatal("expected GetUsers() to propagate repo error as gRPC status")
		}
	})
}

// ── GetUserSummary ───────────────────────────────────────────────────────────

func TestUserGRPCServer_GetUserSummary(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{summary: &domain.UserSummary{Id: "summary-1", Username: "summary_user"}})
		ctx, cancel := testCtx()
		defer cancel()

		resp, err := server.GetUserSummary(ctx, &pb.GetUserSummaryRequest{Id: "summary-1"})
		if err != nil {
			t.Fatalf("GetUserSummary() unexpected error: %v", err)
		}
		if resp == nil || resp.User == nil || resp.User.Id != "summary-1" {
			t.Fatal("expected GetUserSummary() to return the mapped summary response")
		}
	})

	t.Run("error — empty ID rejected", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.GetUserSummary(ctx, &pb.GetUserSummaryRequest{Id: ""})
		if err == nil {
			t.Fatal("expected GetUserSummary() to reject an empty ID")
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{getSummaryErr: errors.New("db error")})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.GetUserSummary(ctx, &pb.GetUserSummaryRequest{Id: "summary-1"})
		if err == nil {
			t.Fatal("expected GetUserSummary() to propagate repo error as gRPC status")
		}
	})
}

// ── GetUsersSummary ──────────────────────────────────────────────────────────

func TestUserGRPCServer_GetUsersSummary(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{summaries: []domain.UserSummary{{Id: "one"}, {Id: "two"}}})
		ctx, cancel := testCtx()
		defer cancel()

		resp, err := server.GetUsersSummary(ctx, &pb.GetUsersSummaryRequest{UserIds: []string{"one", "two"}})
		if err != nil {
			t.Fatalf("GetUsersSummary() unexpected error: %v", err)
		}
		if len(resp.Users) != 2 {
			t.Fatalf("expected two summary values, got %d", len(resp.Users))
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{getSummariesErr: errors.New("db error")})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.GetUsersSummary(ctx, &pb.GetUsersSummaryRequest{UserIds: []string{"one", "two"}})
		if err == nil {
			t.Fatal("expected GetUsersSummary() to propagate repo error as gRPC status")
		}
	})
}

// ── SearchUsers ──────────────────────────────────────────────────────────────

func TestUserGRPCServer_SearchUsers(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{
			searchResult: &domain.UserSearchResult{HasMore: true, Users: []domain.UserSummary{{Id: "search-01"}}},
		})
		ctx, cancel := testCtx()
		defer cancel()

		resp, err := server.SearchUsers(ctx, &pb.SearchUsersRequest{Query: "alice", Page: 1, PageSize: 10})
		if err != nil {
			t.Fatalf("SearchUsers() unexpected error: %v", err)
		}
		if !resp.HasMore || len(resp.Users) != 1 {
			t.Fatal("expected SearchUsers() to return a populated response")
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{searchErr: errors.New("db error")})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.SearchUsers(ctx, &pb.SearchUsersRequest{Query: "alice"})
		if err == nil {
			t.Fatal("expected SearchUsers() to propagate repo error as gRPC status")
		}
	})
}

// ── UserExists ───────────────────────────────────────────────────────────────

func TestUserGRPCServer_UserExists(t *testing.T) {
	t.Run("happy path — user exists", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{existsID: "exists-user"})
		ctx, cancel := testCtx()
		defer cancel()

		resp, err := server.UserExists(ctx, &pb.UserExistsRequest{
			Lookup: &pb.UserExistsRequest_Id{Id: "exists-user"},
		})
		if err != nil {
			t.Fatalf("UserExists() unexpected error: %v", err)
		}
		if resp == nil || resp.Response == nil || !resp.Response.Exists {
			t.Fatal("expected UserExists() to return a positive existence response")
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{existsErr: errors.New("db error")})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.UserExists(ctx, &pb.UserExistsRequest{
			Lookup: &pb.UserExistsRequest_Id{Id: "user-123"},
		})
		if err == nil {
			t.Fatal("expected UserExists() to propagate repo error as gRPC status")
		}
	})
}

// ── UsersExist ───────────────────────────────────────────────────────────────

func TestUserGRPCServer_UsersExist(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{
			existence: []domain.UserExistence{{Id: "a", Exists: true}, {Id: "b", Exists: false}},
		})
		ctx, cancel := testCtx()
		defer cancel()

		resp, err := server.UsersExist(ctx, &pb.UsersExistRequest{UserIds: []string{"a", "b"}})
		if err != nil {
			t.Fatalf("UsersExist() unexpected error: %v", err)
		}
		if len(resp.Users) != 2 {
			t.Fatalf("expected 2 existence entries, got %d", len(resp.Users))
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{usersExistErr: errors.New("db error")})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.UsersExist(ctx, &pb.UsersExistRequest{UserIds: []string{"a", "b"}})
		if err == nil {
			t.Fatal("expected UsersExist() to propagate repo error as gRPC status")
		}
	})
}

// ── GetPrivacySettings ───────────────────────────────────────────────────────

func TestUserGRPCServer_GetPrivacySettings(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{
			privacy: &domain.PrivacySettings{AvatarVisibility: domain.EVERYONE},
		})
		ctx, cancel := testCtx()
		defer cancel()

		resp, err := server.GetPrivacySettings(ctx, &pb.GetPrivacySettingsRequest{UserId: "user-123"})
		if err != nil {
			t.Fatalf("GetPrivacySettings() unexpected error: %v", err)
		}
		if resp == nil || resp.Settings == nil {
			t.Fatal("expected GetPrivacySettings() to return the mapped settings payload")
		}
	})

	t.Run("error — empty user ID", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.GetPrivacySettings(ctx, &pb.GetPrivacySettingsRequest{UserId: ""})
		if err == nil {
			t.Fatal("expected GetPrivacySettings() to reject an empty user ID")
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{privacyErr: errors.New("db error")})
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.GetPrivacySettings(ctx, &pb.GetPrivacySettingsRequest{UserId: "user-123"})
		if err == nil {
			t.Fatal("expected GetPrivacySettings() to propagate repo error as gRPC status")
		}
	})
}

// ── UpdatePrivacySettings ────────────────────────────────────────────────────

func TestUserGRPCServer_UpdatePrivacySettings(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{})
		vis := pb.Visibility_EVERYONE
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.UpdatePrivacySettings(ctx, &pb.UpdatePrivacySettingsRequest{
			UserId:           "user-123",
			AvatarVisibility: &vis,
		})
		if err != nil {
			t.Fatalf("UpdatePrivacySettings() unexpected error: %v", err)
		}
	})

	t.Run("error — repo failure propagated", func(t *testing.T) {
		server := mustNewGRPCServer(t, &grpcFakeRepo{updatePrivacyErr: errors.New("db error")})
		vis := pb.Visibility_FRIENDS
		ctx, cancel := testCtx()
		defer cancel()

		_, err := server.UpdatePrivacySettings(ctx, &pb.UpdatePrivacySettingsRequest{
			UserId:           "user-123",
			AvatarVisibility: &vis,
		})
		if err == nil {
			t.Fatal("expected UpdatePrivacySettings() to propagate repo error as gRPC status")
		}
	})
}
