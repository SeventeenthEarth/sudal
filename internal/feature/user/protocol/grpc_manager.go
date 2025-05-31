package protocol

import (
	"context"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	userv1 "github.com/seventeenthearth/sudal/gen/go/user/v1"
	"github.com/seventeenthearth/sudal/gen/go/user/v1/userv1connect"
	"github.com/seventeenthearth/sudal/internal/feature/user/application"
	"github.com/seventeenthearth/sudal/internal/feature/user/domain/entity"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserManager implements the Connect-go user service
// This handler handles user-related operations including registration, profile retrieval, and updates
type UserManager struct {
	userv1connect.UnimplementedUserServiceHandler
	userService application.UserService
	logger      *zap.Logger
}

// NewUserHandler creates a new user handler with the provided dependencies
// It validates that the logger is not nil, but allows nil userService for testing
func NewUserHandler(userService application.UserService, logger *zap.Logger) *UserManager {
	if logger == nil {
		panic("Logger cannot be nil")
	}

	return &UserManager{
		userService: userService,
		logger:      logger,
	}
}

// RegisterUser implements the RegisterUser RPC method
func (h *UserManager) RegisterUser(
	ctx context.Context,
	req *connect.Request[userv1.RegisterUserRequest],
) (*connect.Response[userv1.RegisterUserResponse], error) {
	h.logger.Info("RegisterUser called",
		zap.String("firebase_uid", req.Msg.FirebaseUid),
		zap.String("display_name", req.Msg.DisplayName),
		zap.String("auth_provider", req.Msg.AuthProvider),
	)

	// Call application service to register user
	user, err := h.userService.RegisterUser(ctx, req.Msg.FirebaseUid, req.Msg.DisplayName, req.Msg.AuthProvider)
	if err != nil {
		h.logger.Error("Failed to register user",
			zap.String("firebase_uid", req.Msg.FirebaseUid),
			zap.Error(err))

		// Map domain errors to gRPC errors
		switch err {
		case entity.ErrUserAlreadyExists:
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		case entity.ErrInvalidFirebaseUID:
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		case entity.ErrInvalidAuthProvider:
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		case entity.ErrInvalidDisplayName:
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	// Convert domain user to proto response
	response := &userv1.RegisterUserResponse{
		UserId: user.ID.String(),
	}

	h.logger.Info("User registered successfully",
		zap.String("user_id", user.ID.String()),
		zap.String("firebase_uid", user.FirebaseUID))

	return connect.NewResponse(response), nil
}

// GetUserProfile implements the GetUserProfile RPC method
func (h *UserManager) GetUserProfile(
	ctx context.Context,
	req *connect.Request[userv1.GetUserProfileRequest],
) (*connect.Response[userv1.UserProfile], error) {
	h.logger.Info("GetUserProfile called",
		zap.String("user_id", req.Msg.UserId),
	)

	// Parse user ID
	userID, err := uuid.Parse(req.Msg.UserId)
	if err != nil {
		h.logger.Error("Invalid user ID format",
			zap.String("user_id", req.Msg.UserId),
			zap.Error(err))
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Call application service to get user profile
	user, err := h.userService.GetUserProfile(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get user profile",
			zap.String("user_id", req.Msg.UserId),
			zap.Error(err))

		// Map domain errors to gRPC errors
		switch err {
		case entity.ErrUserNotFound:
			return nil, connect.NewError(connect.CodeNotFound, err)
		case entity.ErrInvalidUserID:
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	// Convert domain user to proto response
	response := convertUserToProto(user)

	h.logger.Info("User profile retrieved successfully",
		zap.String("user_id", user.ID.String()))

	return connect.NewResponse(response), nil
}

// UpdateUserProfile implements the UpdateUserProfile RPC method
func (h *UserManager) UpdateUserProfile(
	ctx context.Context,
	req *connect.Request[userv1.UpdateUserProfileRequest],
) (*connect.Response[userv1.UpdateUserProfileResponse], error) {
	logFields := []zap.Field{
		zap.String("user_id", req.Msg.UserId),
	}

	if req.Msg.DisplayName != nil {
		logFields = append(logFields, zap.String("display_name", *req.Msg.DisplayName))
	}

	if req.Msg.AvatarUrl != nil {
		logFields = append(logFields, zap.String("avatar_url", *req.Msg.AvatarUrl))
	}

	h.logger.Info("UpdateUserProfile called", logFields...)

	// Parse user ID
	userID, err := uuid.Parse(req.Msg.UserId)
	if err != nil {
		h.logger.Error("Invalid user ID format",
			zap.String("user_id", req.Msg.UserId),
			zap.Error(err))
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Call application service to update user profile
	user, err := h.userService.UpdateUserProfile(ctx, userID, req.Msg.DisplayName, req.Msg.AvatarUrl)
	if err != nil {
		h.logger.Error("Failed to update user profile",
			zap.String("user_id", req.Msg.UserId),
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
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	// Convert domain user to proto response
	response := &userv1.UpdateUserProfileResponse{}

	h.logger.Info("User profile updated successfully",
		zap.String("user_id", user.ID.String()))

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
