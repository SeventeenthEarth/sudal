package connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	userv1 "github.com/seventeenthearth/sudal/gen/go/user/v1"
	"github.com/seventeenthearth/sudal/gen/go/user/v1/userv1connect"
	"github.com/seventeenthearth/sudal/internal/feature/user/domain/repo"
	"go.uber.org/zap"
)

// UserService implements the Connect-go user service
// This service handles user-related operations including registration, profile retrieval, and updates
type UserService struct {
	userv1connect.UnimplementedUserServiceHandler
	repo   repo.UserRepository
	logger *zap.Logger
}

// NewUserService creates a new user service with the provided dependencies
// It validates that the logger is not nil, but allows nil repository for testing
func NewUserService(repository repo.UserRepository, logger *zap.Logger) *UserService {
	if logger == nil {
		panic("Logger cannot be nil")
	}

	return &UserService{
		repo:   repository,
		logger: logger,
	}
}

// RegisterUser implements the RegisterUser RPC method
// Currently returns an unimplemented error as this is scaffolding
func (s *UserService) RegisterUser(
	ctx context.Context,
	req *connect.Request[userv1.RegisterUserRequest],
) (*connect.Response[userv1.RegisterUserResponse], error) {
	s.logger.Info("RegisterUser called",
		zap.String("firebase_uid", req.Msg.FirebaseUid),
		zap.String("display_name", req.Msg.DisplayName),
		zap.String("auth_provider", req.Msg.AuthProvider),
	)

	return nil, connect.NewError(connect.CodeUnimplemented,
		errors.New("user.v1.UserService.RegisterUser is not implemented"))
}

// GetUserProfile implements the GetUserProfile RPC method
// Currently returns an unimplemented error as this is scaffolding
func (s *UserService) GetUserProfile(
	ctx context.Context,
	req *connect.Request[userv1.GetUserProfileRequest],
) (*connect.Response[userv1.UserProfile], error) {
	s.logger.Info("GetUserProfile called",
		zap.String("user_id", req.Msg.UserId),
	)

	return nil, connect.NewError(connect.CodeUnimplemented,
		errors.New("user.v1.UserService.GetUserProfile is not implemented"))
}

// UpdateUserProfile implements the UpdateUserProfile RPC method
// Currently returns an unimplemented error as this is scaffolding
func (s *UserService) UpdateUserProfile(
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

	s.logger.Info("UpdateUserProfile called", logFields...)

	return nil, connect.NewError(connect.CodeUnimplemented,
		errors.New("user.v1.UserService.UpdateUserProfile is not implemented"))
}
