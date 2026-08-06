package handlers

import (
	"net/http"

	"github.com/TheAmgadX/moltaqa-backend/services/api-gateway/internal/middlewares"
	authpb "github.com/TheAmgadX/moltaqa-backend/shared/proto/auth"
	users "github.com/TheAmgadX/moltaqa-backend/shared/proto/users"
	grpc_utils "github.com/TheAmgadX/moltaqa-backend/shared/utils/grpc"
	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	userClient users.UsersServiceClient
	authClient authpb.AuthServiceClient
	// assetClient pb.AssetServiceClient
}

func NewUserHandler(userClient users.UsersServiceClient, authClient authpb.AuthServiceClient) *UserHandler {
	return &UserHandler{
		userClient: userClient,
		authClient: authClient,
	}
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authpb.LoginRequest
	if err := decodeProtoJSON(r.Body, &req); err != nil {
		respondJSONError(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.Login(r.Context(), &req)
	if err != nil {
		respondJSONError(w, grpc_utils.GetGRPCErrorMessage(err), grpc_utils.MapGRPCErrCodesToHttpErrCodes(err))
		return
	}

	respondJSON(w, resp)
}

func (h *UserHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req authpb.VerifyOTPRequest
	if err := decodeProtoJSON(r.Body, &req); err != nil {
		respondJSONError(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.VerifyOTP(r.Context(), &req)
	if err != nil {
		respondJSONError(w, grpc_utils.GetGRPCErrorMessage(err), grpc_utils.MapGRPCErrCodesToHttpErrCodes(err))
		return
	}

	respondJSON(w, resp)
}

func (h *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req authpb.RefreshTokenRequest
	if err := decodeProtoJSON(r.Body, &req); err != nil {
		respondJSONError(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.RefreshToken(r.Context(), &req)
	if err != nil {
		respondJSONError(w, grpc_utils.GetGRPCErrorMessage(err), grpc_utils.MapGRPCErrCodesToHttpErrCodes(err))
		return
	}

	respondJSON(w, resp)
}

// GetUser retrieves a user by id or username or email or phone.
// It returns the user's details if found, or an error if not found or if the request is invalid.
// It supports query parameters for searching by id, username, email, or phone.
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	id := query.Get("id")
	username := query.Get("username")
	email := query.Get("email")
	phone := query.Get("phone")

	req := &users.GetUserRequest{}

	// 1. Map query parameters to the Protobuf oneof lookup field
	switch {
	case id != "":
		req.Lookup = &users.GetUserRequest_Id{Id: id}
	case username != "":
		req.Lookup = &users.GetUserRequest_Username{Username: username}
	case email != "":
		req.Lookup = &users.GetUserRequest_Email{Email: email}
	case phone != "":
		req.Lookup = &users.GetUserRequest_Phone{Phone: phone}
	default:
		respondJSONError(w, "id, username, email, or phone is required", http.StatusBadRequest)
		return
	}

	// 2. Fetch the user from the gRPC microservice
	resp, err := h.userClient.GetUser(r.Context(), req)
	if err != nil {
		respondJSONError(w, grpc_utils.GetGRPCErrorMessage(err), grpc_utils.MapGRPCErrCodesToHttpErrCodes(err))
		return
	}

	fetchedUser := resp.User

	// 3. PRIVACY & AUTHORIZATION ENFORCEMENT
	// Get the caller's ID from the context (injected by Auth middleware)
	// If the user is a guest (unauthenticated), callerID will be empty.
	callerID, _ := r.Context().Value(middlewares.UserIDKey).(string)

	// If the caller is NOT the owner of the profile.
	if callerID != fetchedUser.Id {
		// TODO: Here we will apply privacy rules based on the user's privacy settings
	}

	// 4. Return the sanitized user data
	respondJSON(w, fetchedUser)
}

// UpdateUserProfile updates authenticated user's profile info.
func (h *UserHandler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	callerID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || callerID == "" {
		respondJSONError(w, "unauthorized access", http.StatusUnauthorized)
		return
	}

	var req users.UpdateUserRequest
	if err := decodeProtoJSON(r.Body, &req); err != nil {
		respondJSONError(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	req.Id = callerID

	resp, err := h.userClient.UpdateUser(r.Context(), &req)
	if err != nil {
		respondJSONError(w, grpc_utils.GetGRPCErrorMessage(err), grpc_utils.MapGRPCErrCodesToHttpErrCodes(err))
		return
	}

	respondJSON(w, resp)
}

// DeleteUserAccount requests deletion for the authenticated account.
func (h *UserHandler) DeleteUserAccount(w http.ResponseWriter, r *http.Request) {
	callerID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || callerID == "" {
		respondJSONError(w, "unauthorized access", http.StatusUnauthorized)
		return
	}

	var req authpb.DeleteAccountRequest
	if err := decodeProtoJSON(r.Body, &req); err != nil {
		respondJSONError(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.DeleteAccount(r.Context(), &req)
	if err != nil {
		respondJSONError(w, grpc_utils.GetGRPCErrorMessage(err), grpc_utils.MapGRPCErrCodesToHttpErrCodes(err))
		return
	}

	respondJSON(w, resp)
}

// Restore requests account restoration for authenticated user.
func (h *UserHandler) Restore(w http.ResponseWriter, r *http.Request) {
	callerID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || callerID == "" {
		respondJSONError(w, "unauthorized access", http.StatusUnauthorized)
		return
	}

	var req authpb.RestoreAccountRequest
	if err := decodeProtoJSON(r.Body, &req); err != nil {
		respondJSONError(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	resp, err := h.authClient.RestoreAccount(r.Context(), &req)
	if err != nil {
		respondJSONError(w, grpc_utils.GetGRPCErrorMessage(err), grpc_utils.MapGRPCErrCodesToHttpErrCodes(err))
		return
	}

	respondJSON(w, resp)
}

// UploadProfileImage updates the authenticated user's profile image.
func (h *UserHandler) UploadProfileImage(w http.ResponseWriter, r *http.Request) {
	// Implement it with the asset service
}

// GetPrivacySettings fetches current user's privacy settings.
func (h *UserHandler) GetPrivacySettings(w http.ResponseWriter, r *http.Request) {
	callerID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || callerID == "" {
		respondJSONError(w, "unauthorized access", http.StatusUnauthorized)
		return
	}

	req := &users.GetPrivacySettingsRequest{UserId: callerID}

	resp, err := h.userClient.GetPrivacySettings(r.Context(), req)
	if err != nil {
		respondJSONError(w, grpc_utils.GetGRPCErrorMessage(err), grpc_utils.MapGRPCErrCodesToHttpErrCodes(err))
		return
	}

	respondJSON(w, resp.Settings)
}

// UpdatePrivacySettings updates privacy settings for authenticated user.
func (h *UserHandler) UpdatePrivacySettings(w http.ResponseWriter, r *http.Request) {
	callerID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || callerID == "" {
		respondJSONError(w, "unauthorized access", http.StatusUnauthorized)
		return
	}

	var req users.UpdatePrivacySettingsRequest
	if err := decodeProtoJSON(r.Body, &req); err != nil {
		respondJSONError(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	req.UserId = callerID
	resp, err := h.userClient.UpdatePrivacySettings(r.Context(), &req)
	if err != nil {
		respondJSONError(w, grpc_utils.GetGRPCErrorMessage(err), grpc_utils.MapGRPCErrCodesToHttpErrCodes(err))
		return
	}

	respondJSON(w, resp)
}

// SearchUsers performs paginated search for users.
func (h *UserHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	var req users.SearchUsersRequest
	if err := decodeProtoJSON(r.Body, &req); err != nil {
		respondJSONError(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Apply defaults if not provided in JSON
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	resp, err := h.userClient.SearchUsers(r.Context(), &req)
	if err != nil {
		respondJSONError(w, grpc_utils.GetGRPCErrorMessage(err), grpc_utils.MapGRPCErrCodesToHttpErrCodes(err))
		return
	}

	respondJSON(w, resp)
}

// GetUserSummary retrieves a single user's public summary card.
func (h *UserHandler) GetUserSummary(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		id = r.URL.Query().Get("id")
	}

	if id == "" {
		respondJSONError(w, "user id parameter is required", http.StatusBadRequest)
		return
	}

	req := &users.GetUserSummaryRequest{Id: id}

	resp, err := h.userClient.GetUserSummary(r.Context(), req)
	if err != nil {
		respondJSONError(w, grpc_utils.GetGRPCErrorMessage(err), grpc_utils.MapGRPCErrCodesToHttpErrCodes(err))
		return
	}

	respondJSON(w, resp.User)
}

// GetUsersSummary retrieves batch public summaries for multiple users.
func (h *UserHandler) GetUsersSummary(w http.ResponseWriter, r *http.Request) {
	var req users.GetUsersSummaryRequest
	if err := decodeProtoJSON(r.Body, &req); err != nil {
		respondJSONError(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	if len(req.UserIds) == 0 {
		respondJSONError(w, "user_ids list cannot be empty", http.StatusBadRequest)
		return
	}

	resp, err := h.userClient.GetUsersSummary(r.Context(), &req)
	if err != nil {
		respondJSONError(w, grpc_utils.GetGRPCErrorMessage(err), grpc_utils.MapGRPCErrCodesToHttpErrCodes(err))
		return
	}

	respondJSON(w, resp.Users)
}

// RegisterContact links a phone or email to the authenticated user.
func (h *UserHandler) RegisterContact(w http.ResponseWriter, r *http.Request) {
	callerID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || callerID == "" {
		respondJSONError(w, "unauthorized access", http.StatusUnauthorized)
		return
	}

	var req users.RegisterContactRequest
	if err := decodeProtoJSON(r.Body, &req); err != nil {
		respondJSONError(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	req.UserId = callerID

	// Check that at least one value was set in the oneof field
	if req.ContactType == nil {
		respondJSONError(w, "email or phone contact field is required", http.StatusBadRequest)
		return
	}

	resp, err := h.userClient.RegisterContact(r.Context(), &req)
	if err != nil {
		respondJSONError(w, grpc_utils.GetGRPCErrorMessage(err), grpc_utils.MapGRPCErrCodesToHttpErrCodes(err))
		return
	}

	respondJSON(w, resp)
}

// UserExists checks single account existence by id, username, email, or phone.
func (h *UserHandler) UserExists(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	id := query.Get("id")
	username := query.Get("username")
	email := query.Get("email")
	phone := query.Get("phone")

	req := &users.UserExistsRequest{}

	switch {
	case id != "":
		req.Lookup = &users.UserExistsRequest_Id{Id: id}
	case username != "":
		req.Lookup = &users.UserExistsRequest_Username{Username: username}
	case email != "":
		req.Lookup = &users.UserExistsRequest_Email{Email: email}
	case phone != "":
		req.Lookup = &users.UserExistsRequest_Phone{Phone: phone}
	default:
		respondJSONError(w, "id, username, email, or phone is required", http.StatusBadRequest)
		return
	}

	resp, err := h.userClient.UserExists(r.Context(), req)
	if err != nil {
		respondJSONError(w, grpc_utils.GetGRPCErrorMessage(err), grpc_utils.MapGRPCErrCodesToHttpErrCodes(err))
		return
	}

	respondJSON(w, resp.Response)
}
