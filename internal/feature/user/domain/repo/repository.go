package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/seventeenthearth/sudal/internal/feature/user/domain/entity"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../../mocks/mock_user_repository.go -package=mocks -mock_names=UserRepository=MockUserRepository github.com/seventeenthearth/sudal/internal/feature/user/domain/repo UserRepository

// UserRepository defines the protocol for user data access operations
// This protocol abstracts the data layer and supports both PostgreSQL and Redis implementations
// following the Repository Pattern to maintain clean separation between domain and data layers.
//
// Implementation Strategy:
// - Write Operations (Create/Update): Write to PostgreSQL first, then update/invalidate Redis cache
// - Read Operations: Attempt Redis first (cache), fallback to PostgreSQL on cache miss
//
// Error Handling:
// - Use domain-specific sentinel errors (entity.ErrUserNotFound, entity.ErrUserAlreadyExists, etc.)
// - Wrap infrastructure errors with appropriate context
// - Ensure consistent error types across different implementations
type UserRepository interface {
	// Create creates a new user in the system
	// The user must have a valid FirebaseUID and AuthProvider
	// Returns entity.ErrUserAlreadyExists if a user with the same FirebaseUID already exists
	// Returns entity.ErrInvalidFirebaseUID if the FirebaseUID is empty or invalid
	// Returns entity.ErrInvalidAuthProvider if the AuthProvider is empty or invalid
	Create(ctx context.Context, user *entity.User) (*entity.User, error)

	// GetByID retrieves a user by their unique ID
	// Returns entity.ErrUserNotFound if no user exists with the given ID
	// Returns entity.ErrInvalidUserID if the provided ID is invalid (e.g., not a valid UUID)
	GetByID(ctx context.Context, userID uuid.UUID) (*entity.User, error)

	// GetByFirebaseUID retrieves a user by their Firebase authentication UID
	// This is commonly used during authentication flows to find existing users
	// Returns entity.ErrUserNotFound if no user exists with the given Firebase UID
	// Returns entity.ErrInvalidFirebaseUID if the provided Firebase UID is empty or invalid
	GetByFirebaseUID(ctx context.Context, firebaseUID string) (*entity.User, error)

	// Update updates an existing user's information
	// The user ID must exist in the system
	// Only non-zero/non-nil fields will be updated (partial updates supported)
	// The UpdatedAt timestamp will be automatically set to the current time
	// Returns entity.ErrUserNotFound if no user exists with the given ID
	// Returns entity.ErrInvalidDisplayName if the display name violates constraints
	Update(ctx context.Context, user *entity.User) (*entity.User, error)

	// UpdateCandyBalance updates a user's candy balance by the specified amount
	// Positive amounts add candy, negative amounts deduct candy
	// Returns entity.ErrUserNotFound if no user exists with the given ID
	// Returns entity.ErrInsufficientBalance if deducting more candy than available
	// Returns entity.ErrInvalidAmount if the amount would result in a negative balance
	UpdateCandyBalance(ctx context.Context, userID uuid.UUID, amount int) (*entity.User, error)

	// Delete removes a user from the system (soft delete recommended)
	// This operation should also clean up related data (quiz results, etc.)
	// Returns entity.ErrUserNotFound if no user exists with the given ID
	// Note: Consider implementing soft delete for data retention and audit purposes
	Delete(ctx context.Context, userID uuid.UUID) error

	// Exists checks if a user exists with the given Firebase UID
	// This is useful for registration flows to check if a user already exists
	// Returns true if the user exists, false otherwise
	// Returns an error only if there's a system/database error
	Exists(ctx context.Context, firebaseUID string) (bool, error)

	// List retrieves a paginated list of users
	// offset: number of users to skip (for pagination)
	// limit: maximum number of users to return (should have a reasonable maximum)
	// Returns an empty slice if no users are found (not an error)
	// This method is primarily for administrative purposes
	List(ctx context.Context, offset, limit int) ([]*entity.User, error)

	// Count returns the total number of users in the system
	// This is useful for pagination calculations and administrative dashboards
	Count(ctx context.Context) (int64, error)
}
