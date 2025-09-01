package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/seventeenthearth/sudal/internal/feature/user/domain/entity"
	"github.com/seventeenthearth/sudal/internal/feature/user/domain/repo"
	"github.com/seventeenthearth/sudal/internal/infrastructure/repository/postgres"
	ssql "github.com/seventeenthearth/sudal/internal/service/sql"
	ssqlpg "github.com/seventeenthearth/sudal/internal/service/sql/postgres"
)

// userRepoImpl is the PostgreSQL implementation of the UserRepository protocol
// It embeds the shared postgres.Repository to leverage common database functionality
// while implementing user-specific data access operations.
//
// This implementation follows the Repository Pattern and provides:
// - PostgreSQL-specific user data repository
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
// and returns an instance that satisfies the user.UserRepository protocol.
//
// Parameters:
//   - db: The database connection pool (*sql.DB) for executing queries
//   - logger: Structured logger for recording repository operations and errors
//
// Returns:
//   - repo.UserRepository: A repository instance that implements the user domain protocol
//
// Example Usage:
//
//	userRepo := repo.NewUserRepoImpl(dbConnection, logger)
//	user, err := userRepo.GetByID(ctx, userID)
func NewUserRepoImpl(db *sql.DB, logger *zap.Logger) repo.UserRepository {
	exec, _ := ssqlpg.NewFromDB(db)
	return NewUserRepoWithExecutor(exec, logger)
}

// NewUserRepoWithExecutor creates a repository using the minimal SQL executor interface.
func NewUserRepoWithExecutor(exec ssql.Executor, logger *zap.Logger) repo.UserRepository {
	return &userRepoImpl{Repository: postgres.NewRepositoryWithExecutor(exec, logger)}
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

	row := r.GetExecutor().QueryRowContext(ctx, query,
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

	// Execute SELECT query with userID (exclude soft-deleted users)
	query := `
		SELECT id, firebase_uid, display_name, avatar_url, candy_balance, auth_provider, created_at, updated_at
		FROM sudal.users
		WHERE id = $1 AND deleted_at IS NULL`

	row := r.GetExecutor().QueryRowContext(ctx, query, userID)

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
	// Validate firebaseUID parameter
	if firebaseUID == "" {
		return nil, entity.ErrInvalidFirebaseUID
	}

	// Execute SELECT query with firebaseUID (exclude soft-deleted users)
	query := `
		SELECT id, firebase_uid, display_name, avatar_url, candy_balance, auth_provider, created_at, updated_at
		FROM sudal.users
		WHERE firebase_uid = $1 AND deleted_at IS NULL`

	row := r.GetExecutor().QueryRowContext(ctx, query, firebaseUID)

	user := &entity.User{}
	err := row.Scan(
		&user.ID, &user.FirebaseUID, &user.DisplayName,
		&user.AvatarURL, &user.CandyBalance, &user.AuthProvider,
		&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrUserNotFound
		}
		r.GetLogger().Error("Failed to get user by Firebase UID",
			zap.String("firebase_uid", firebaseUID),
			zap.Error(err))
		return nil, err
	}

	return user, nil
}

// Update updates an existing user's information
// Only non-zero fields in the provided user will be updated
// Returns entity.ErrUserNotFound if no user exists with the given ID
// Returns entity.ErrInvalidUserID if the provided user ID is invalid
func (r *userRepoImpl) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	// Validate user parameter and ID
	if user == nil {
		return nil, entity.ErrInvalidUserID
	}
	if user.ID == uuid.Nil {
		return nil, entity.ErrInvalidUserID
	}

	// Validate display name if provided
	if user.DisplayName != nil && !entity.IsValidDisplayName(*user.DisplayName) {
		return nil, entity.ErrInvalidDisplayName
	}

	// Build dynamic UPDATE query for non-zero/non-nil fields
	setParts := []string{"updated_at = CURRENT_TIMESTAMP"}
	args := []any{}
	argIndex := 1

	if user.DisplayName != nil {
		setParts = append(setParts, fmt.Sprintf("display_name = $%d", argIndex))
		args = append(args, user.DisplayName)
		argIndex++
	}

	if user.AvatarURL != nil {
		setParts = append(setParts, fmt.Sprintf("avatar_url = $%d", argIndex))
		args = append(args, user.AvatarURL)
		argIndex++
	}

	if user.CandyBalance != 0 {
		setParts = append(setParts, fmt.Sprintf("candy_balance = $%d", argIndex))
		args = append(args, user.CandyBalance)
		argIndex++
	}

	if user.AuthProvider != "" {
		setParts = append(setParts, fmt.Sprintf("auth_provider = $%d", argIndex))
		args = append(args, user.AuthProvider)
		argIndex++
	}

	// Add user ID as the final argument for WHERE clause
	args = append(args, user.ID)

	query := fmt.Sprintf(`
		UPDATE sudal.users
		SET %s
		WHERE id = $%d
		RETURNING id, firebase_uid, display_name, avatar_url, candy_balance, auth_provider, created_at, updated_at`,
		strings.Join(setParts, ", "), argIndex)

	row := r.GetExecutor().QueryRowContext(ctx, query, args...)

	updatedUser := &entity.User{}
	err := row.Scan(
		&updatedUser.ID, &updatedUser.FirebaseUID, &updatedUser.DisplayName,
		&updatedUser.AvatarURL, &updatedUser.CandyBalance, &updatedUser.AuthProvider,
		&updatedUser.CreatedAt, &updatedUser.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, entity.ErrUserNotFound
		}
		r.GetLogger().Error("Failed to update user",
			zap.String("user_id", user.ID.String()),
			zap.Error(err))
		return nil, err
	}

	r.GetLogger().Info("User updated successfully",
		zap.String("user_id", updatedUser.ID.String()))

	return updatedUser, nil
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
	// Validate userID parameter
	if userID == uuid.Nil {
		return entity.ErrInvalidUserID
	}

	// Perform soft delete by setting deleted_at timestamp
	// Only update if the user exists and is not already soft-deleted
	query := `
		UPDATE sudal.users
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.GetExecutor().ExecContext(ctx, query, userID)
	if err != nil {
		r.GetLogger().Error("Failed to soft delete user",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return err
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.GetLogger().Error("Failed to get rows affected for user deletion",
			zap.String("user_id", userID.String()),
			zap.Error(err))
		return err
	}

	if rowsAffected == 0 {
		// No rows were updated, which means either:
		// 1. User doesn't exist, or
		// 2. User is already soft-deleted
		// In both cases, we return ErrUserNotFound
		return entity.ErrUserNotFound
	}

	r.GetLogger().Info("User soft deleted successfully",
		zap.String("user_id", userID.String()))

	return nil
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

	// Execute COUNT query (exclude soft-deleted users)
	query := `SELECT COUNT(1) FROM sudal.users WHERE firebase_uid = $1 AND deleted_at IS NULL`

	row := r.GetExecutor().QueryRowContext(ctx, query, firebaseUID)

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
	// Validate offset and limit parameters
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 50 // Default reasonable limit
	}
	if limit > 1000 {
		limit = 1000 // Maximum reasonable limit
	}

	// Execute SELECT query with LIMIT and OFFSET (exclude soft-deleted users)
	query := `
		SELECT id, firebase_uid, display_name, avatar_url, candy_balance, auth_provider, created_at, updated_at
		FROM sudal.users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.GetExecutor().QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.GetLogger().Error("Failed to list users",
			zap.Int("offset", offset),
			zap.Int("limit", limit),
			zap.Error(err))
		return nil, err
	}
	defer rows.Close() // nolint:errcheck

	var users []*entity.User
	for rows.Next() {
		user := &entity.User{}
		err := rows.Scan(
			&user.ID, &user.FirebaseUID, &user.DisplayName,
			&user.AvatarURL, &user.CandyBalance, &user.AuthProvider,
			&user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			r.GetLogger().Error("Failed to scan user row",
				zap.Error(err))
			return nil, err
		}
		users = append(users, user)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		r.GetLogger().Error("Error iterating over user rows",
			zap.Error(err))
		return nil, err
	}

	r.GetLogger().Debug("Users listed successfully",
		zap.Int("count", len(users)),
		zap.Int("offset", offset),
		zap.Int("limit", limit))

	return users, nil
}

// Count returns the total number of users in the system
// This is useful for pagination calculations and administrative dashboards
func (r *userRepoImpl) Count(ctx context.Context) (int64, error) {
	// Execute COUNT query (exclude soft-deleted users)
	query := `SELECT COUNT(*) FROM sudal.users WHERE deleted_at IS NULL`

	row := r.GetExecutor().QueryRowContext(ctx, query)

	var count int64
	err := row.Scan(&count)
	if err != nil {
		r.GetLogger().Error("Failed to count users",
			zap.Error(err))
		return 0, err
	}

	r.GetLogger().Debug("User count retrieved successfully",
		zap.Int64("count", count))

	return count, nil
}
