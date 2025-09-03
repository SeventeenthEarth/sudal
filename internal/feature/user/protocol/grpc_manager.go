package protocol

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	userv1 "github.com/seventeenthearth/sudal/gen/go/user/v1"
	"github.com/seventeenthearth/sudal/gen/go/user/v1/userv1connect"
	"github.com/seventeenthearth/sudal/internal/feature/user/application"
	"github.com/seventeenthearth/sudal/internal/feature/user/domain/entity"
	"github.com/seventeenthearth/sudal/internal/infrastructure/middleware"
	"github.com/seventeenthearth/sudal/internal/service/authutil"
	"github.com/seventeenthearth/sudal/internal/service/firebaseauth"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserManager implements the Connect-go user service
// This handler handles user-related operations including registration, profile retrieval, and updates
type UserManager struct {
	userv1connect.UnimplementedUserServiceHandler
	userService   application.UserService
	tokenVerifier firebaseauth.TokenVerifier
	logger        *zap.Logger
}

// NewUserManager creates a new user handler with the provided dependencies
// It validates that the logger is not nil, but allows nil userService for testing
func NewUserManager(userService application.UserService, tokenVerifier firebaseauth.TokenVerifier, logger *zap.Logger) *UserManager {
	if logger == nil {
		panic("Logger cannot be nil")
	}

	return &UserManager{
		userService:   userService,
		tokenVerifier: tokenVerifier,
		logger:        logger,
	}
}

// RegisterUser implements the RegisterUser RPC method
// This method verifies the Firebase ID token and creates a new user account
func (h *UserManager) RegisterUser(
	ctx context.Context,
	req *connect.Request[userv1.RegisterUserRequest],
) (*connect.Response[userv1.RegisterUserResponse], error) {
	h.logger.Info("RegisterUser called",
		zap.String("firebase_uid", req.Msg.FirebaseUid),
		zap.String("display_name", req.Msg.DisplayName),
		zap.String("auth_provider", req.Msg.AuthProvider),
	)

	// Extract and verify Firebase ID token from Authorization header
	authHeader := req.Header().Get("Authorization")
	if authHeader == "" {
		h.logger.Warn("Missing Authorization header for RegisterUser")
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("missing authorization header"))
	}

	// Extract Bearer token
	token, err := authutil.ExtractBearerToken(authHeader)
	if err != nil {
		h.logger.Warn("Invalid Authorization header format for RegisterUser", zap.Error(err))
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid authorization header format"))
	}

	// Verify Firebase ID token
	if h.tokenVerifier == nil {
		h.logger.Warn("Token verifier is not configured")
		// Treat as server misconfiguration
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("internal server error"))
	}
	uid, provider, err := h.tokenVerifier.Verify(ctx, token)
	if err != nil {
		h.logger.Warn("Firebase token verification failed for RegisterUser", zap.Error(err))
		// Do not leak verification error details to clients
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authentication failed"))
	}

	// Verify that the Firebase UID in the token matches the request
	if uid != req.Msg.FirebaseUid {
		h.logger.Warn("Firebase UID mismatch",
			zap.String("token_firebase_uid", uid),
			zap.String("request_firebase_uid", req.Msg.FirebaseUid))
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("firebase UID mismatch"))
	}

	// Ensure user exists in application layer.
	// Pass initial display name via context to keep business logic in the application layer.
	if req.Msg.DisplayName != "" {
		ctx = application.WithInitialDisplayName(ctx, req.Msg.DisplayName)
	}
	user, err := h.userService.EnsureUserByFirebaseUID(ctx, uid, provider)
	if err != nil {
		h.logger.Error("Failed to ensure user existence",
			zap.String("firebase_uid", uid), zap.Error(err))
		// Do not expose internal error details
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("internal server error"))
	}

	response := &userv1.RegisterUserResponse{UserId: user.ID.String(), Success: true}
	h.logger.Info("User registered successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("firebase_uid", uid))

	return connect.NewResponse(response), nil
}

// GetUserProfile implements the GetUserProfile RPC method
// This method retrieves the authenticated user's profile from the request context
func (h *UserManager) GetUserProfile(
	ctx context.Context,
	req *connect.Request[userv1.GetUserProfileRequest],
) (*connect.Response[userv1.UserProfile], error) {
	// Get authenticated user from context (injected by authentication middleware)
	user, err := middleware.GetAuthenticatedUser(ctx)
	if err != nil {
		h.logger.Error("Failed to get authenticated user from context", zap.Error(err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("internal server error"))
	}

	h.logger.Info("GetUserProfile called",
		zap.String("authenticated_user_id", user.ID.String()),
		zap.String("requested_user_id", req.Msg.UserId),
	)

	// Verify that the requested user ID matches the authenticated user
	if user.ID.String() != req.Msg.UserId {
		h.logger.Warn("User ID mismatch - user trying to access another user's profile",
			zap.String("authenticated_user_id", user.ID.String()),
			zap.String("requested_user_id", req.Msg.UserId))
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("cannot access another user's profile"))
	}

	// Convert domain user to proto response (user is already from database via middleware)
	response := convertUserToProto(user)

	h.logger.Info("User profile retrieved successfully",
		zap.String("user_id", user.ID.String()))

	return connect.NewResponse(response), nil
}

// UpdateUserProfile implements the UpdateUserProfile RPC method
// This method updates the authenticated user's profile using data from the request context
func (h *UserManager) UpdateUserProfile(
	ctx context.Context,
	req *connect.Request[userv1.UpdateUserProfileRequest],
) (*connect.Response[userv1.UpdateUserProfileResponse], error) {
	// Get authenticated user from context (injected by authentication middleware)
	user, err := middleware.GetAuthenticatedUser(ctx)
	if err != nil {
		h.logger.Error("Failed to get authenticated user from context", zap.Error(err))
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("internal server error"))
	}

	logFields := []zap.Field{
		zap.String("authenticated_user_id", user.ID.String()),
		zap.String("requested_user_id", req.Msg.UserId),
	}

	if req.Msg.DisplayName != nil {
		logFields = append(logFields, zap.String("display_name", *req.Msg.DisplayName))
	}

	if req.Msg.AvatarUrl != nil {
		logFields = append(logFields, zap.String("avatar_url", *req.Msg.AvatarUrl))
	}

	h.logger.Info("UpdateUserProfile called", logFields...)

	// Verify that the requested user ID matches the authenticated user
	if user.ID.String() != req.Msg.UserId {
		h.logger.Warn("User ID mismatch - user trying to update another user's profile",
			zap.String("authenticated_user_id", user.ID.String()),
			zap.String("requested_user_id", req.Msg.UserId))
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("cannot update another user's profile"))
	}

	// Call application service to update user profile
	updatedUser, err := h.userService.UpdateUserProfile(ctx, user.ID, req.Msg.DisplayName, req.Msg.AvatarUrl)
	if err != nil {
		h.logger.Error("Failed to update user profile",
			zap.String("user_id", user.ID.String()),
			zap.Error(err))

		// Map domain errors to gRPC errors
		switch err {
		case entity.ErrUserNotFound:
			return nil, connect.NewError(connect.CodeNotFound, err)
		case entity.ErrInvalidUserID:
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		case entity.ErrInvalidDisplayName:
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		default:
			// Do not expose internal error details to clients
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("internal server error"))
		}
	}

	// Convert domain user to proto response
	response := &userv1.UpdateUserProfileResponse{
		Success: true,
	}

	h.logger.Info("User profile updated successfully",
		zap.String("user_id", updatedUser.ID.String()))

	return connect.NewResponse(response), nil
}

// convertUserToProto converts a domain User entity to a proto UserProfile
func convertUserToProto(user *entity.User) *userv1.UserProfile {
	profile := &userv1.UserProfile{
		UserId:       user.ID.String(),
		CandyBalance: int32(user.CandyBalance),
		CreatedAt:    timestamppb.New(user.CreatedAt),
	}

	if user.DisplayName != nil {
		profile.DisplayName = *user.DisplayName
	}

	if user.AvatarURL != nil {
		profile.AvatarUrl = *user.AvatarURL
	}

	return profile
}

// NewUserHandler creates a new UserManager instance for Wire dependency injection
// This function is used by Wire to create the UserManager with all required dependencies
func NewUserHandler(userService application.UserService, tokenVerifier firebaseauth.TokenVerifier, logger *zap.Logger) *UserManager {
	return NewUserManager(userService, tokenVerifier, logger)
}
