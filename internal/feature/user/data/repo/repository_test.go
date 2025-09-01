package repo

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/seventeenthearth/sudal/internal/feature/user/domain/entity"
)

// setupTestRepo creates a test repository with mocked database
func setupTestRepo(t *testing.T) (*userRepoImpl, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	logger := zap.NewNop() // Use no-op logger for tests
	repo := NewUserRepoWithExecutor(db, logger).(*userRepoImpl)

	cleanup := func() {
		db.Close() // nolint:errcheck
	}

	return repo, mock, cleanup
}

// createTestUser creates a test user entity
func createTestUser() *entity.User {
	userID := uuid.New()
	displayName := "Test User"
	avatarURL := "https://example.com/avatar.jpg"
	now := time.Now().UTC()

	return &entity.User{
		ID:           userID,
		FirebaseUID:  "firebase_test_uid_123",
		DisplayName:  &displayName,
		AvatarURL:    &avatarURL,
		CandyBalance: 100,
		AuthProvider: "google",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func TestUserRepoImpl_Create_Success(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	user := createTestUser()

	// Mock the Exists check (should return false for new user)
	mock.ExpectQuery(`SELECT COUNT\(1\) FROM sudal\.users WHERE firebase_uid = \$1`).
		WithArgs(user.FirebaseUID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock the INSERT query
	mock.ExpectQuery(`INSERT INTO sudal\.users \(id, firebase_uid, display_name, avatar_url, candy_balance, auth_provider, created_at, updated_at\)`).
		WithArgs(user.ID, user.FirebaseUID, user.DisplayName, user.AvatarURL, user.CandyBalance, user.AuthProvider, user.CreatedAt, user.UpdatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id", "firebase_uid", "display_name", "avatar_url", "candy_balance", "auth_provider", "created_at", "updated_at"}).
			AddRow(user.ID, user.FirebaseUID, user.DisplayName, user.AvatarURL, user.CandyBalance, user.AuthProvider, user.CreatedAt, user.UpdatedAt))

	// Act
	result, err := repo.Create(context.Background(), user)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, user.FirebaseUID, result.FirebaseUID)
	assert.Equal(t, user.DisplayName, result.DisplayName)
	assert.Equal(t, user.CandyBalance, result.CandyBalance)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepoImpl_Create_UserAlreadyExists(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	user := createTestUser()

	// Mock the Exists check (should return true for existing user)
	mock.ExpectQuery(`SELECT COUNT\(1\) FROM sudal\.users WHERE firebase_uid = \$1`).
		WithArgs(user.FirebaseUID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Act
	result, err := repo.Create(context.Background(), user)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, entity.ErrUserAlreadyExists, err)
	assert.Nil(t, result)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepoImpl_Create_InvalidInput(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	tests := []struct {
		name        string
		user        *entity.User
		expectedErr error
	}{
		{
			name:        "nil user",
			user:        nil,
			expectedErr: entity.ErrInvalidUserID,
		},
		{
			name: "empty firebase UID",
			user: &entity.User{
				ID:           uuid.New(),
				FirebaseUID:  "",
				AuthProvider: "google",
			},
			expectedErr: entity.ErrInvalidFirebaseUID,
		},
		{
			name: "empty auth provider",
			user: &entity.User{
				ID:           uuid.New(),
				FirebaseUID:  "firebase_uid_123",
				AuthProvider: "",
			},
			expectedErr: entity.ErrInvalidAuthProvider,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result, err := repo.Create(context.Background(), tt.user)

			// Assert
			assert.Error(t, err)
			assert.Equal(t, tt.expectedErr, err)
			assert.Nil(t, result)
		})
	}

	// Verify all expectations were met (should be none for validation errors)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepoImpl_GetByID_Success(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	user := createTestUser()

	// Mock the SELECT query
	mock.ExpectQuery(`SELECT id, firebase_uid, display_name, avatar_url, candy_balance, auth_provider, created_at, updated_at FROM sudal\.users WHERE id = \$1`).
		WithArgs(user.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "firebase_uid", "display_name", "avatar_url", "candy_balance", "auth_provider", "created_at", "updated_at"}).
			AddRow(user.ID, user.FirebaseUID, user.DisplayName, user.AvatarURL, user.CandyBalance, user.AuthProvider, user.CreatedAt, user.UpdatedAt))

	// Act
	result, err := repo.GetByID(context.Background(), user.ID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, user.FirebaseUID, result.FirebaseUID)
	assert.Equal(t, user.DisplayName, result.DisplayName)
	assert.Equal(t, user.CandyBalance, result.CandyBalance)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepoImpl_GetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	userID := uuid.New()

	// Mock the SELECT query to return no rows
	mock.ExpectQuery(`SELECT id, firebase_uid, display_name, avatar_url, candy_balance, auth_provider, created_at, updated_at FROM sudal\.users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	// Act
	result, err := repo.GetByID(context.Background(), userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, entity.ErrUserNotFound, err) // Critical: verify error wrapping
	assert.Nil(t, result)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepoImpl_GetByID_InvalidUserID(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	// Act
	result, err := repo.GetByID(context.Background(), uuid.Nil)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, entity.ErrInvalidUserID, err)
	assert.Nil(t, result)

	// Verify all expectations were met (should be none for validation errors)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepoImpl_Exists_True(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	firebaseUID := "firebase_test_uid_123"

	// Mock the COUNT query to return 1
	mock.ExpectQuery(`SELECT COUNT\(1\) FROM sudal\.users WHERE firebase_uid = \$1`).
		WithArgs(firebaseUID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	// Act
	exists, err := repo.Exists(context.Background(), firebaseUID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, exists)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepoImpl_Exists_False(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	firebaseUID := "nonexistent_firebase_uid"

	// Mock the COUNT query to return 0
	mock.ExpectQuery(`SELECT COUNT\(1\) FROM sudal\.users WHERE firebase_uid = \$1`).
		WithArgs(firebaseUID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Act
	exists, err := repo.Exists(context.Background(), firebaseUID)

	// Assert
	assert.NoError(t, err)
	assert.False(t, exists)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepoImpl_Exists_InvalidFirebaseUID(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	// Act
	exists, err := repo.Exists(context.Background(), "")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, entity.ErrInvalidFirebaseUID, err)
	assert.False(t, exists)

	// Verify all expectations were met (should be none for validation errors)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepoImpl_GetByID_DatabaseError(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	userID := uuid.New()
	expectedError := sql.ErrConnDone

	// Mock the SELECT query to return a database error
	mock.ExpectQuery(`SELECT id, firebase_uid, display_name, avatar_url, candy_balance, auth_provider, created_at, updated_at FROM sudal\.users WHERE id = \$1`).
		WithArgs(userID).
		WillReturnError(expectedError)

	// Act
	result, err := repo.GetByID(context.Background(), userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err) // Should return the original database error
	assert.Nil(t, result)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepoImpl_Exists_DatabaseError(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	firebaseUID := "firebase_test_uid_123"
	expectedError := sql.ErrConnDone

	// Mock the COUNT query to return a database error
	mock.ExpectQuery(`SELECT COUNT\(1\) FROM sudal\.users WHERE firebase_uid = \$1`).
		WithArgs(firebaseUID).
		WillReturnError(expectedError)

	// Act
	exists, err := repo.Exists(context.Background(), firebaseUID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err) // Should return the original database error
	assert.False(t, exists)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepoImpl_Create_DatabaseError(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	user := createTestUser()
	expectedError := sql.ErrConnDone

	// Mock the Exists check (should return false for new user)
	mock.ExpectQuery(`SELECT COUNT\(1\) FROM sudal\.users WHERE firebase_uid = \$1`).
		WithArgs(user.FirebaseUID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock the INSERT query to return a database error
	mock.ExpectQuery(`INSERT INTO sudal\.users \(id, firebase_uid, display_name, avatar_url, candy_balance, auth_provider, created_at, updated_at\)`).
		WithArgs(user.ID, user.FirebaseUID, user.DisplayName, user.AvatarURL, user.CandyBalance, user.AuthProvider, user.CreatedAt, user.UpdatedAt).
		WillReturnError(expectedError)

	// Act
	result, err := repo.Create(context.Background(), user)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err) // Should return the original database error
	assert.Nil(t, result)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepoImpl_Create_ExistsCheckError(t *testing.T) {
	repo, mock, cleanup := setupTestRepo(t)
	defer cleanup()

	user := createTestUser()
	expectedError := sql.ErrConnDone

	// Mock the Exists check to return a database error
	mock.ExpectQuery(`SELECT COUNT\(1\) FROM sudal\.users WHERE firebase_uid = \$1`).
		WithArgs(user.FirebaseUID).
		WillReturnError(expectedError)

	// Act
	result, err := repo.Create(context.Background(), user)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err) // Should return the original database error
	assert.Nil(t, result)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
