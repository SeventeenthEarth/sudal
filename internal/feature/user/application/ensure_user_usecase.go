package application

import (
	"context"

	"github.com/seventeenthearth/sudal/internal/feature/user/domain/entity"
	"github.com/seventeenthearth/sudal/internal/feature/user/domain/repo"
)

// EnsureUserUseCase defines the protocol to ensure a user exists for a Firebase UID.
// If the user does not exist, it creates one with the given provider.
type EnsureUserUseCase interface {
	Execute(ctx context.Context, firebaseUID, authProvider string) (*entity.User, error)
}

type ensureUserUseCase struct {
	repo repo.UserRepository
}

// NewEnsureUserUseCase creates a new EnsureUserUseCase implementation.
func NewEnsureUserUseCase(repository repo.UserRepository) EnsureUserUseCase {
	return &ensureUserUseCase{repo: repository}
}

func (uc *ensureUserUseCase) Execute(ctx context.Context, firebaseUID, authProvider string) (*entity.User, error) {
	// Try to find existing user
	user, err := uc.repo.GetByFirebaseUID(ctx, firebaseUID)
	if err == nil {
		return user, nil
	}

	if err != nil && err != entity.ErrUserNotFound {
		return nil, err
	}

	// Create a new user if not found
	newUser := entity.NewUser(firebaseUID, authProvider)
	return uc.repo.Create(ctx, newUser)
}
