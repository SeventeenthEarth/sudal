package application

import (
	"context"
	"strings"

	"github.com/seventeenthearth/sudal/internal/feature/user/domain/entity"
	"github.com/seventeenthearth/sudal/internal/feature/user/domain/repo"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../mocks/mock_register_user_usecase.go -package=mocks github.com/seventeenthearth/sudal/internal/feature/user/application RegisterUserUseCase

// RegisterUserUseCase defines the protocol for user registration functionality
type RegisterUserUseCase interface {
	Execute(ctx context.Context, firebaseUID, displayName, authProvider string) (*entity.User, error)
}

// registerUserUseCase implements the RegisterUserUseCase protocol
type registerUserUseCase struct {
	repo repo.UserRepository
}

// NewRegisterUserUseCase creates a new register user use case
func NewRegisterUserUseCase(repository repo.UserRepository) RegisterUserUseCase {
	return &registerUserUseCase{
		repo: repository,
	}
}

// Execute performs user registration with business logic validation
func (uc *registerUserUseCase) Execute(ctx context.Context, firebaseUID, displayName, authProvider string) (*entity.User, error) {
	// Validate input parameters
	if strings.TrimSpace(firebaseUID) == "" {
		return nil, entity.ErrInvalidFirebaseUID
	}

	if strings.TrimSpace(authProvider) == "" {
		return nil, entity.ErrInvalidAuthProvider
	}

	// Create new user entity with business rules
	user := entity.NewUser(firebaseUID, authProvider)

	// Set display name if provided and valid
	if strings.TrimSpace(displayName) != "" {
		if !entity.IsValidDisplayName(displayName) {
			return nil, entity.ErrInvalidDisplayName
		}
		user.UpdateDisplayName(displayName)
	}

	// Create user in repository
	createdUser, err := uc.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}
