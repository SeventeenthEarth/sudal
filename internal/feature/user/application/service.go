package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/seventeenthearth/sudal/internal/feature/user/domain/entity"
	"github.com/seventeenthearth/sudal/internal/feature/user/domain/repo"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../mocks/mock_user_service.go -package=mocks github.com/seventeenthearth/sudal/internal/feature/user/application UserService

// UserService defines the user service protocol
// This service acts as a facade for user-related business operations
type UserService interface {
	// RegisterUser creates a new user account after Firebase authentication
	// Returns the created user or an error if registration fails
	RegisterUser(ctx context.Context, firebaseUID, displayName, authProvider string) (*entity.User, error)

	// EnsureUserByFirebaseUID ensures a user exists for the given Firebase UID
	// If not exists, it creates a new one with the given auth provider
	EnsureUserByFirebaseUID(ctx context.Context, firebaseUID, authProvider string) (*entity.User, error)

	// GetUserProfile retrieves a user's profile by their ID
	// Returns the user profile or an error if not found
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*entity.User, error)

	// UpdateUserProfile updates a user's profile information
	// Only provided fields will be updated (partial update)
	// Returns the updated user or an error if update fails
	UpdateUserProfile(ctx context.Context, userID uuid.UUID, displayName, avatarURL *string) (*entity.User, error)
}

// userServiceImpl is the implementation of the user service
// It acts as a facade for the individual use cases
type userServiceImpl struct {
	registerUserUseCase      RegisterUserUseCase
	ensureUserUseCase        EnsureUserUseCase
	getUserProfileUseCase    GetUserProfileUseCase
	updateUserProfileUseCase UpdateUserProfileUseCase
}

// NewService creates a new user service
func NewService(repository repo.UserRepository) UserService {
	return &userServiceImpl{
		registerUserUseCase:      NewRegisterUserUseCase(repository),
		ensureUserUseCase:        NewEnsureUserUseCase(repository),
		getUserProfileUseCase:    NewGetUserProfileUseCase(repository),
		updateUserProfileUseCase: NewUpdateUserProfileUseCase(repository),
	}
}

// RegisterUser creates a new user account after Firebase authentication
func (s *userServiceImpl) RegisterUser(ctx context.Context, firebaseUID, displayName, authProvider string) (*entity.User, error) {
	return s.registerUserUseCase.Execute(ctx, firebaseUID, displayName, authProvider)
}

// EnsureUserByFirebaseUID returns an existing user or creates a new one
func (s *userServiceImpl) EnsureUserByFirebaseUID(ctx context.Context, firebaseUID, authProvider string) (*entity.User, error) {
	return s.ensureUserUseCase.Execute(ctx, firebaseUID, authProvider)
}

// GetUserProfile retrieves a user's profile by their ID
func (s *userServiceImpl) GetUserProfile(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	return s.getUserProfileUseCase.Execute(ctx, userID)
}

// UpdateUserProfile updates a user's profile information
func (s *userServiceImpl) UpdateUserProfile(ctx context.Context, userID uuid.UUID, displayName, avatarURL *string) (*entity.User, error) {
	return s.updateUserProfileUseCase.Execute(ctx, userID, displayName, avatarURL)
}
