package application

import (
	"context"
	"errors"

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
	// Check if an initial display name was provided via context
	initName, hasInitName := initialDisplayNameFromContext(ctx)

	// Try to find existing user
	user, err := uc.repo.GetByFirebaseUID(ctx, firebaseUID)
	if err == nil {
		// If user exists and has no display name yet, and an initial name is provided, update it
		if hasInitName && user.DisplayName == nil {
			user.UpdateDisplayName(initName)
			return uc.repo.Update(ctx, user)
		}
		return user, nil
	}

	// If the error is anything other than "not found", it's an unexpected error.
	if !errors.Is(err, entity.ErrUserNotFound) {
		return nil, err
	}

	// Create a new user if not found
	newUser := entity.NewUser(firebaseUID, authProvider)
	if hasInitName {
		// Let repository/domain validate constraints (will return ErrInvalidDisplayName if invalid)
		newUser.UpdateDisplayName(initName)
	}
	created, cerr := uc.repo.Create(ctx, newUser)
	if cerr == nil {
		return created, nil
	}
	// Handle potential race: another concurrent request created the user
	if errors.Is(cerr, entity.ErrUserAlreadyExists) {
		existing, gerr := uc.repo.GetByFirebaseUID(ctx, firebaseUID)
		if gerr != nil {
			return nil, gerr
		}
		if hasInitName && existing.DisplayName == nil {
			existing.UpdateDisplayName(initName)
			return uc.repo.Update(ctx, existing)
		}
		return existing, nil
	}
	return nil, cerr
}
