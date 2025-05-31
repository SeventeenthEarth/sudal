package repo

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/seventeenthearth/sudal/internal/feature/user/domain/entity"
	"github.com/seventeenthearth/sudal/internal/feature/user/domain/repo"
	"github.com/seventeenthearth/sudal/internal/infrastructure/persistence/postgres"
)

// userRepoImpl is the PostgreSQL implementation of the UserRepository interface
// It embeds the shared postgres.Repository to leverage common database functionality
// while implementing user-specific data access operations.
//
// This implementation follows the Repository Pattern and provides:
// - PostgreSQL-specific user data persistence
// - Transaction support through the embedded base repository
// - Standardized error handling using domain-specific errors
// - Structured logging for all database operations
type userRepoImpl struct {
	// Embed the shared PostgreSQL repository component
	// This provides access to database connection, transaction management,
	// and common repository functionality
	*postgres.Repository
}

// NewUserRepoImpl creates a new PostgreSQL-based user repository implementation
// This constructor initializes the repository with the shared PostgreSQL infrastructure
// and returns an instance that satisfies the user.UserRepository interface.
//
// Parameters:
//   - db: The database connection pool (*sql.DB) for executing queries
//   - logger: Structured logger for recording repository operations and errors
//
// Returns:
//   - repo.UserRepository: A repository instance that implements the user domain interface
//
// Example Usage:
//
//	userRepo := repo.NewUserRepoImpl(dbConnection, logger)
//	user, err := userRepo.GetByID(ctx, userID)
func NewUserRepoImpl(db *sql.DB, logger *zap.Logger) repo.UserRepository {
	return &userRepoImpl{
		Repository: postgres.NewRepository(db, logger),
	}
}

// Create creates a new user in the system
// The user must have a valid FirebaseUID and AuthProvider
// Returns entity.ErrUserAlreadyExists if a user with the same FirebaseUID already exists
// Returns entity.ErrInvalidFirebaseUID if the FirebaseUID is empty or invalid
// Returns entity.ErrInvalidAuthProvider if the AuthProvider is empty or invalid
func (r *userRepoImpl) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	// Validate input parameters
	if user == nil {
		return nil, entity.ErrInvalidUserID
	}
	if user.FirebaseUID == "" {
		return nil, entity.ErrInvalidFirebaseUID
	}
	if user.AuthProvider == "" {
		return nil, entity.ErrInvalidAuthProvider
	}

	// Check if user already exists with the same FirebaseUID
	exists, err := r.Exists(ctx, user.FirebaseUID)
	if err != nil {
		r.GetLogger().Error("Failed to check user existence",
			zap.String("firebase_uid", user.FirebaseUID),
			zap.Error(err))
		return nil, err
	}
	if exists {
		return nil, entity.ErrUserAlreadyExists
	}

	// Insert new user record
	query := `
		INSERT INTO sudal.users (id, firebase_uid, display_name, avatar_url, candy_balance, auth_provider, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, firebase_uid, display_name, avatar_url, candy_balance, auth_provider, created_at, updated_at`

	db := r.GetDB().(interface {
		QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	})

	row := db.QueryRowContext(ctx, query,
		user.ID, user.FirebaseUID, user.DisplayName, user.AvatarURL,
		user.CandyBalance, user.AuthProvider, user.CreatedAt, user.UpdatedAt)

	createdUser := &entity.User{}
	err = row.Scan(
		&createdUser.ID, &createdUser.FirebaseUID, &createdUser.DisplayName,
		&createdUser.AvatarURL, &createdUser.CandyBalance, &createdUser.AuthProvider,
		&createdUser.CreatedAt, &createdUser.UpdatedAt)

	if err != nil {
		r.GetLogger().Error("Failed to create user",
			zap.String("firebase_uid", user.FirebaseUID),
			zap.Error(err))
		return nil, err
	}

	r.GetLogger().Info("User created successfully",
		zap.String("user_id", createdUser.ID.String()),
		zap.String("firebase_uid", createdUser.FirebaseUID))

	return createdUser, nil
}

// GetByID retrieves a user by their unique ID
// Returns entity.ErrUserNotFound if no user exists with the given ID
// Returns entity.ErrInvalidUserID if the provided ID is invalid (e.g., not a valid UUID)
func (r *userRepoImpl) GetByID(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	// Validate userID parameter
	if userID == uuid.Nil {
		return nil, entity.ErrInvalidUserID
	}

	// Execute SELECT query with userID
	query := `
		SELECT id, firebase_uid, display_name, avatar_url, candy_balance, auth_provider, created_at, updated_at
		FROM sudal.users
		WHERE id = $1`

	db := r.GetDB().(interface {
		QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	})

	row := db.QueryRowContext(ctx, query, userID)

	user := &entity.User{}
	err := row.Scan(
		&user.ID, &user.FirebaseUID, &user.DisplayName,
		&user.AvatarURL, &user.CandyBalance, &user.AuthProvider,
		&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrUserNotFound
		}
		r.GetLogger().Error("Failed to get user by ID",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return nil, err
	}

	return user, nil
}

// GetByFirebaseUID retrieves a user by their Firebase authentication UID
// Returns entity.ErrUserNotFound if no user exists with the given Firebase UID
// Returns entity.ErrInvalidFirebaseUID if the provided Firebase UID is empty or invalid
func (r *userRepoImpl) GetByFirebaseUID(ctx context.Context, firebaseUID string) (*entity.User, error) {
	// TODO: Implement user retrieval by Firebase UID
	// This should include:
	// 1. Validate firebaseUID parameter
	// 2. Execute SELECT query with firebaseUID
	// 3. Scan result into User entity
	// 4. Handle not found case appropriately
	panic("not implemented")
}

// Update updates an existing user's information
// Only non-zero fields in the provided user will be updated
// Returns entity.ErrUserNotFound if no user exists with the given ID
// Returns entity.ErrInvalidUserID if the provided user ID is invalid
func (r *userRepoImpl) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	// TODO: Implement user update logic
	// This should include:
	// 1. Validate user parameter and ID
	// 2. Build dynamic UPDATE query for non-zero fields
	// 3. Execute update with optimistic locking if needed
	// 4. Return updated user with new timestamp
	panic("not implemented")
}

// UpdateCandyBalance updates a user's candy balance by the specified amount
// Positive amounts add candy, negative amounts deduct candy
// Returns entity.ErrUserNotFound if no user exists with the given ID
// Returns entity.ErrInsufficientBalance if deducting more candy than available
// Returns entity.ErrInvalidAmount if the amount would result in a negative balance
func (r *userRepoImpl) UpdateCandyBalance(ctx context.Context, userID uuid.UUID, amount int) (*entity.User, error) {
	// TODO: Implement candy balance update logic
	// This should include:
	// 1. Validate userID parameter
	// 2. Check current balance if deducting
	// 3. Execute UPDATE query for candy_balance
	// 4. Return updated user entity
	panic("not implemented")
}

// Delete removes a user from the system (soft delete recommended)
// This operation should also clean up related data (quiz results, etc.)
// Returns entity.ErrUserNotFound if no user exists with the given ID
// Note: Consider implementing soft delete for data retention and audit purposes
func (r *userRepoImpl) Delete(ctx context.Context, userID uuid.UUID) error {
	// TODO: Implement user deletion (preferably soft delete)
	// This should include:
	// 1. Validate userID parameter
	// 2. Check if user exists
	// 3. Perform soft delete (set deleted_at timestamp)
	// 4. Consider cleanup of related data
	panic("not implemented")
}

// Exists checks if a user exists with the given Firebase UID
// This is useful for registration flows to check if a user already exists
// Returns true if the user exists, false otherwise
// Returns an error only if there's a system/database error
func (r *userRepoImpl) Exists(ctx context.Context, firebaseUID string) (bool, error) {
	// Validate firebaseUID parameter
	if firebaseUID == "" {
		return false, entity.ErrInvalidFirebaseUID
	}

	// Execute COUNT query
	query := `SELECT COUNT(1) FROM sudal.users WHERE firebase_uid = $1`

	db := r.GetDB().(interface {
		QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	})

	row := db.QueryRowContext(ctx, query, firebaseUID)

	var count int
	err := row.Scan(&count)
	if err != nil {
		r.GetLogger().Error("Failed to check user existence",
			zap.String("firebase_uid", firebaseUID),
			zap.Error(err))
		return false, err
	}

	return count > 0, nil
}

// List retrieves a paginated list of users
// offset: number of users to skip (for pagination)
// limit: maximum number of users to return (should have a reasonable maximum)
// Returns an empty slice if no users are found (not an error)
// This method is primarily for administrative purposes
func (r *userRepoImpl) List(ctx context.Context, offset, limit int) ([]*entity.User, error) {
	// TODO: Implement paginated user listing
	// This should include:
	// 1. Validate offset and limit parameters
	// 2. Execute SELECT query with LIMIT and OFFSET
	// 3. Scan results into User entities slice
	// 4. Return empty slice if no results (not an error)
	panic("not implemented")
}

// Count returns the total number of users in the system
// This is useful for pagination calculations and administrative dashboards
func (r *userRepoImpl) Count(ctx context.Context) (int64, error) {
	// TODO: Implement user count
	// This should include:
	// 1. Execute COUNT(*) query
	// 2. Return total count as int64
	// 3. Handle database errors appropriately
	panic("not implemented")
}
