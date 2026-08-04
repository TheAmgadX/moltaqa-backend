package grpc_test

import (
	"context"
	"testing"
	"time"

	"github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/domain"
	grpcpkg "github.com/TheAmgadX/moltaqa-backend/services/user-service/internal/infrastructure/grpc"
	pb "github.com/TheAmgadX/moltaqa-backend/shared/proto/users"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ── lookupRecorderRepo ────────────────────────────────────────────────────────
// A specialised fake repo that records the lookup passed to Get() and Exists()
// so mapper_test.go can assert the correct domain.Lookup type was constructed.

type lookupRecorderRepo struct {
	grpcFakeRepo
	lastLookup       domain.Lookup
	lastExistsLookup domain.Lookup
}

// newLookupRecorder returns a recorder with a non-nil fallback user so that
// server handlers that call mapUserToProto on the response don't nil-deref.
func newLookupRecorder() *lookupRecorderRepo {
	return &lookupRecorderRepo{
		grpcFakeRepo: grpcFakeRepo{
			user: &domain.User{Email: "recorder@example.com"},
		},
	}
}

func (r *lookupRecorderRepo) Get(_ context.Context, lookup domain.Lookup) (*domain.User, error) {
	r.lastLookup = lookup
	return r.grpcFakeRepo.user, r.grpcFakeRepo.getErr
}

func (r *lookupRecorderRepo) Exists(_ context.Context, lookup domain.Lookup) (string, error) {
	r.lastExistsLookup = lookup
	return r.grpcFakeRepo.existsID, r.grpcFakeRepo.existsErr
}

// ── MapProtoVisibilityToDomain ────────────────────────────────────────────────

func TestMapProtoVisibilityToDomain(t *testing.T) {
	tests := []struct {
		name  string
		input pb.Visibility
		want  domain.Visibility
	}{
		{"EVERYONE", pb.Visibility_EVERYONE, domain.EVERYONE},
		{"FRIENDS", pb.Visibility_FRIENDS, domain.FRIENDS},
		{"CONTACTS", pb.Visibility_CONTACTS, domain.CONTACTS},
		{"NOBODY", pb.Visibility_NOBODY, domain.NOBODY},
		{"unknown value returns empty string", pb.Visibility(999), ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := grpcpkg.MapProtoVisibilityToDomain(tc.input)
			if got != tc.want {
				t.Fatalf("want %q, got %q", tc.want, got)
			}
		})
	}
}

// ── GetUser — all lookup types ────────────────────────────────────────────────

func TestGetUserRequest_AllLookupTypes(t *testing.T) {
	tests := []struct {
		name      string
		req       *pb.GetUserRequest
		wantType  domain.LookupType
		wantValue string
	}{
		{
			name:      "lookup by ID",
			req:       &pb.GetUserRequest{Lookup: &pb.GetUserRequest_Id{Id: "user-id-123"}},
			wantType:  domain.LookUpId,
			wantValue: "user-id-123",
		},
		{
			name:      "lookup by username",
			req:       &pb.GetUserRequest{Lookup: &pb.GetUserRequest_Username{Username: "john_doe"}},
			wantType:  domain.LookupUsername,
			wantValue: "john_doe",
		},
		{
			name:      "lookup by email",
			req:       &pb.GetUserRequest{Lookup: &pb.GetUserRequest_Email{Email: "john@example.com"}},
			wantType:  domain.LookupEmail,
			wantValue: "john@example.com",
		},
		{
			name:      "lookup by phone",
			req:       &pb.GetUserRequest{Lookup: &pb.GetUserRequest_Phone{Phone: "+12065550199"}},
			wantType:  domain.LookupPhone,
			wantValue: "+12065550199",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			recorder := newLookupRecorder()
			recServer := mustNewGRPCServerWithRepo(t, recorder)
			ctx, cancel := testCtx()
			defer cancel()

			_, _ = recServer.GetUser(ctx, tc.req)

			if recorder.lastLookup.Type != tc.wantType {
				t.Fatalf("want lookup type %v, got %v", tc.wantType, recorder.lastLookup.Type)
			}
			if recorder.lastLookup.Value != tc.wantValue {
				t.Fatalf("want lookup value %q, got %q", tc.wantValue, recorder.lastLookup.Value)
			}
		})
	}
}

// ── UserExists — all lookup types ─────────────────────────────────────────────

func TestUserExistsRequest_AllLookupTypes(t *testing.T) {
	tests := []struct {
		name      string
		req       *pb.UserExistsRequest
		wantType  domain.LookupType
		wantValue string
	}{
		{
			name:      "exists by ID",
			req:       &pb.UserExistsRequest{Lookup: &pb.UserExistsRequest_Id{Id: "user-id-123"}},
			wantType:  domain.LookUpId,
			wantValue: "user-id-123",
		},
		{
			name:      "exists by username",
			req:       &pb.UserExistsRequest{Lookup: &pb.UserExistsRequest_Username{Username: "john_doe"}},
			wantType:  domain.LookupUsername,
			wantValue: "john_doe",
		},
		{
			name:      "exists by email",
			req:       &pb.UserExistsRequest{Lookup: &pb.UserExistsRequest_Email{Email: "john@example.com"}},
			wantType:  domain.LookupEmail,
			wantValue: "john@example.com",
		},
		{
			name:      "exists by phone",
			req:       &pb.UserExistsRequest{Lookup: &pb.UserExistsRequest_Phone{Phone: "+12065550199"}},
			wantType:  domain.LookupPhone,
			wantValue: "+12065550199",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			recorder := newLookupRecorder()
			recServer := mustNewGRPCServerWithRepo(t, recorder)
			ctx, cancel := testCtx()
			defer cancel()

			_, _ = recServer.UserExists(ctx, tc.req)

			if recorder.lastExistsLookup.Type != tc.wantType {
				t.Fatalf("want lookup type %v, got %v", tc.wantType, recorder.lastExistsLookup.Type)
			}
			if recorder.lastExistsLookup.Value != tc.wantValue {
				t.Fatalf("want lookup value %q, got %q", tc.wantValue, recorder.lastExistsLookup.Value)
			}
		})
	}
}

// ── Account badge — all values ────────────────────────────────────────────────

func TestAccountBadge_AllValues(t *testing.T) {
	// Exercises mapAccountBadgeToDomain via UpdateUser and mapAccountBadgeToProto
	// via mapUserToProto inside CreateUser/GetUser response paths.
	badges := []struct {
		name  string
		proto pb.AccountBadge
	}{
		{"UNVERIFIED", pb.AccountBadge_UNVERIFIED},
		{"BLUE_BADGE", pb.AccountBadge_BLUE_BADGE},
		{"GOLD_BADGE", pb.AccountBadge_GOLD_BADGE},
		{"SILVER_BADGE", pb.AccountBadge_SILVER_BADGE},
	}

	for _, tc := range badges {
		t.Run(tc.name, func(t *testing.T) {
			badgeVal := tc.proto
			server := mustNewGRPCServer(t, &grpcFakeRepo{})
			ctx, cancel := testCtx()
			defer cancel()

			_, err := server.UpdateUser(ctx, &pb.UpdateUserRequest{
				Id:           "11111111-1111-1111-1111-111111111111",
				AccountBadge: &badgeVal,
			})
			if err != nil {
				t.Fatalf("UpdateUser() with badge %v returned error: %v", tc.proto, err)
			}
		})
	}
}

// ── RegisterContact — phone branch ───────────────────────────────────────────

func TestRegisterContact_PhoneBranch(t *testing.T) {
	server := mustNewGRPCServer(t, &grpcFakeRepo{})
	ctx, cancel := testCtx()
	defer cancel()

	_, err := server.RegisterContact(ctx, &pb.RegisterContactRequest{
		UserId:      "user-123",
		ContactType: &pb.RegisterContactRequest_Phone{Phone: "+12065550199"},
	})
	if err != nil {
		t.Fatalf("RegisterContact() phone branch unexpected error: %v", err)
	}
}

// ── UpdatePrivacySettings — all visibility fields ─────────────────────────────

func TestUpdatePrivacySettings_AllVisibilityFields(t *testing.T) {
	server := mustNewGRPCServer(t, &grpcFakeRepo{})
	ctx, cancel := testCtx()
	defer cancel()

	everyone := pb.Visibility_EVERYONE
	friends := pb.Visibility_FRIENDS
	contacts := pb.Visibility_CONTACTS
	nobody := pb.Visibility_NOBODY
	readReceipts := true
	findByUsername := false

	_, err := server.UpdatePrivacySettings(ctx, &pb.UpdatePrivacySettingsRequest{
		UserId:              "user-123",
		AvatarVisibility:    &everyone,
		PhoneVisibility:     &friends,
		EmailVisibility:     &contacts,
		LastSeenVisibility:  &nobody,
		ReadReceiptsEnabled: &readReceipts,
		FindByUsername:      &findByUsername,
	})
	if err != nil {
		t.Fatalf("UpdatePrivacySettings() all fields unexpected error: %v", err)
	}
}

// ── mapTimeToDomain — nil branch ──────────────────────────────────────────────

func TestUpdateUser_NilBirthDate(t *testing.T) {
	server := mustNewGRPCServer(t, &grpcFakeRepo{})
	ctx, cancel := testCtx()
	defer cancel()

	_, err := server.UpdateUser(ctx, &pb.UpdateUserRequest{
		Id:        "11111111-1111-1111-1111-111111111111",
		BirthDate: nil,
	})
	if err != nil {
		t.Fatalf("UpdateUser() nil birth date unexpected error: %v", err)
	}
}

// ── mapTimeToProto — zero time returns nil ────────────────────────────────────

func TestMapUserToProto_ZeroTimeReturnsNilTimestamp(t *testing.T) {
	server := mustNewGRPCServer(t, &grpcFakeRepo{})
	ctx, cancel := testCtx()
	defer cancel()

	resp, err := server.CreateUser(ctx, &pb.CreateUserRequest{
		Contact: &pb.CreateUserRequest_Email{Email: "zero@example.com"},
	})
	if err != nil {
		t.Fatalf("CreateUser() unexpected error: %v", err)
	}
	if resp.User.BirthDate != nil {
		t.Fatalf("expected nil BirthDate for zero-time user, got %v", resp.User.BirthDate)
	}
}

// ── mapTimeToProto — non-zero time returns timestamp ─────────────────────────

func TestUpdateUser_NonZeroBirthDate(t *testing.T) {
	server := mustNewGRPCServer(t, &grpcFakeRepo{})
	ctx, cancel := testCtx()
	defer cancel()

	past := time.Now().Add(-365 * 24 * time.Hour)
	_, err := server.UpdateUser(ctx, &pb.UpdateUserRequest{
		Id:        "11111111-1111-1111-1111-111111111111",
		BirthDate: timestamppb.New(past),
	})
	if err != nil {
		t.Fatalf("UpdateUser() non-zero birth date unexpected error: %v", err)
	}
}
