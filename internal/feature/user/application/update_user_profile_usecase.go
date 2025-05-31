package application

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/seventeenthearth/sudal/internal/feature/user/domain/entity"
	"github.com/seventeenthearth/sudal/internal/feature/user/domain/repo"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../mocks/mock_update_user_profile_usecase.go -package=mocks github.com/seventeenthearth/sudal/internal/feature/user/application UpdateUserProfileUseCase

// UpdateUserProfileUseCase defines the interface for user profile update functionality
type UpdateUserProfileUseCase interface {
	Execute(ctx context.Context, userID uuid.UUID, displayName, avatarURL *string) (*entity.User, error)
}

// updateUserProfileUseCase implements the UpdateUserProfileUseCase interface
type updateUserProfileUseCase struct {
	repo repo.UserRepository
}

// NewUpdateUserProfileUseCase creates a new update user profile use case
func NewUpdateUserProfileUseCase(repository repo.UserRepository) UpdateUserProfileUseCase {
	return &updateUserProfileUseCase{
		repo: repository,
	}
}

// Execute updates a user profile with business logic validation
func (uc *updateUserProfileUseCase) Execute(ctx context.Context, userID uuid.UUID, displayName, avatarURL *string) (*entity.User, error) {
	// Validate input parameters
	if userID == uuid.Nil {
		return nil, entity.ErrInvalidUserID
	}

	// Get existing user to ensure it exists
	existingUser, err := uc.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Create a copy for updates
	updatedUser := *existingUser

	// Validate and update display name if provided
	if displayName != nil {
		trimmedDisplayName := strings.TrimSpace(*displayName)
		if trimmedDisplayName == "" {
			// Empty string means remove display name
			updatedUser.DisplayName = nil
		} else {
			if !entity.IsValidDisplayName(trimmedDisplayName) {
				return nil, entity.ErrInvalidDisplayName
			}
			updatedUser.UpdateDisplayName(trimmedDisplayName)
		}
	}

	// Update avatar URL if provided
	if avatarURL != nil {
		trimmedAvatarURL := strings.TrimSpace(*avatarURL)
		if trimmedAvatarURL == "" {
			// Empty string means remove avatar URL
			updatedUser.AvatarURL = nil
		} else {
			updatedUser.UpdateAvatarURL(trimmedAvatarURL)
		}
	}

	// Update user in repository
	result, err := uc.repo.Update(ctx, &updatedUser)
	if err != nil {
		return nil, err
	}

	return result, nil
}
