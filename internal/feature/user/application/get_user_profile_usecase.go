package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/seventeenthearth/sudal/internal/feature/user/domain/entity"
	"github.com/seventeenthearth/sudal/internal/feature/user/domain/repo"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../mocks/mock_get_user_profile_usecase.go -package=mocks github.com/seventeenthearth/sudal/internal/feature/user/application GetUserProfileUseCase

// GetUserProfileUseCase defines the interface for user profile retrieval functionality
type GetUserProfileUseCase interface {
	Execute(ctx context.Context, userID uuid.UUID) (*entity.User, error)
}

// getUserProfileUseCase implements the GetUserProfileUseCase interface
type getUserProfileUseCase struct {
	repo repo.UserRepository
}

// NewGetUserProfileUseCase creates a new get user profile use case
func NewGetUserProfileUseCase(repository repo.UserRepository) GetUserProfileUseCase {
	return &getUserProfileUseCase{
		repo: repository,
	}
}

// Execute retrieves a user profile by ID with business logic validation
func (uc *getUserProfileUseCase) Execute(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	// Validate input parameters
	if userID == uuid.Nil {
		return nil, entity.ErrInvalidUserID
	}

	// Retrieve user from repository
	user, err := uc.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
