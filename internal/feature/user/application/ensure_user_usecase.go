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
	if err != nil {
		if !errors.Is(err, entity.ErrUserNotFound) {
			// If the error is anything other than "not found", it's an unexpected error.
			return nil, err
		}

		// User not found, create a new one.
		newUser := entity.NewUser(firebaseUID, authProvider)
		if hasInitName {
			// Let repository/domain validate constraints (will return ErrInvalidDisplayName if invalid)
			newUser.UpdateDisplayName(initName)
		}
		created, cerr := uc.repo.Create(ctx, newUser)
		if cerr == nil {
			return created, nil
		}

		// Handle potential race: another concurrent request created the user.
		if !errors.Is(cerr, entity.ErrUserAlreadyExists) {
			return nil, cerr
		}

		// Fetch the user that was just created by the other request.
		user, err = uc.repo.GetByFirebaseUID(ctx, firebaseUID)
		if err != nil {
			return nil, err // Should not happen if ErrUserAlreadyExists was returned, but handle it.
		}
	}

	// At this point, 'user' is a valid user entity.
	// If an initial name is provided and the user doesn't have one, update it.
	if hasInitName && user.DisplayName == nil {
		user.UpdateDisplayName(initName)
		return uc.repo.Update(ctx, user)
	}

	return user, nil
}
