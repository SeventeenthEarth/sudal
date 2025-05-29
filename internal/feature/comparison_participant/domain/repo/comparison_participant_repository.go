package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/seventeenthearth/sudal/internal/feature/comparison_participant/domain/entity"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../../mocks/mock_comparison_participant_repository.go -package=mocks -mock_names=ComparisonParticipantRepository=MockComparisonParticipantRepository github.com/seventeenthearth/sudal/internal/feature/comparison_participant/domain/repo ComparisonParticipantRepository

// ComparisonParticipantRepository defines the interface for comparison participant data access operations
// This interface abstracts the data layer and supports both PostgreSQL and Redis implementations
// following the Repository Pattern to maintain clean separation between domain and data layers.
//
// Implementation Strategy:
// - Write Operations (Create/Update): Write to PostgreSQL first, then update/invalidate Redis cache
// - Read Operations: Attempt Redis first (cache), fallback to PostgreSQL on cache miss
//
// Error Handling:
// - Use domain-specific sentinel errors (entity.ErrComparisonParticipantNotFound, entity.ErrComparisonParticipantAlreadyExists, etc.)
// - Wrap infrastructure errors with appropriate context
// - Ensure consistent error types across different implementations
type ComparisonParticipantRepository interface {
	// Create adds a user to a comparison event
	// The user must not already be participating in the comparison
	// Returns entity.ErrComparisonParticipantDuplicateUser if the user is already participating
	// Returns entity.ErrComparisonParticipantInvalidComparisonID if the comparison doesn't exist
	// Returns entity.ErrComparisonParticipantInvalidQuizResultID if the quiz result is not valid for this comparison
	Create(ctx context.Context, participant *entity.ComparisonParticipant) (*entity.ComparisonParticipant, error)

	// GetByID retrieves a specific participant by their ID
	// Returns entity.ErrComparisonParticipantNotFound if the participant doesn't exist
	GetByID(ctx context.Context, id int64) (*entity.ComparisonParticipant, error)

	// GetByComparisonID retrieves all participants for a specific comparison
	// Returns an empty slice if no participants are found (not an error)
	GetByComparisonID(ctx context.Context, comparisonID int64) ([]*entity.ComparisonParticipant, error)

	// GetByUserID retrieves all participations for a specific user
	// page: page number (1-based)
	// pageSize: maximum number of participants to return per page
	// Returns participants ordered by created_at DESC, total count, and error
	GetByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*entity.ComparisonParticipant, int64, error)

	// GetByComparisonAndUser retrieves a specific participation record
	// Returns entity.ErrComparisonParticipantNotFound if the participation doesn't exist
	GetByComparisonAndUser(ctx context.Context, comparisonID int64, userID uuid.UUID) (*entity.ComparisonParticipant, error)

	// Update updates an existing participant's information (mainly personal memo)
	// Returns entity.ErrComparisonParticipantNotFound if the participant doesn't exist
	Update(ctx context.Context, participant *entity.ComparisonParticipant) (*entity.ComparisonParticipant, error)

	// UpdatePersonalMemo updates a participant's personal memo
	// Returns entity.ErrComparisonParticipantNotFound if the participant doesn't exist
	UpdatePersonalMemo(ctx context.Context, comparisonID int64, userID uuid.UUID, memo string) error

	// Delete removes a user from a comparison event
	// Returns entity.ErrComparisonParticipantNotFound if the participant doesn't exist
	Delete(ctx context.Context, id int64) error

	// DeleteByComparisonAndUser removes a specific participation record
	// Returns entity.ErrComparisonParticipantNotFound if the participation doesn't exist
	DeleteByComparisonAndUser(ctx context.Context, comparisonID int64, userID uuid.UUID) error

	// DeleteByComparisonID removes all participants for a specific comparison
	// This is useful when deleting an entire comparison
	// Returns the number of participants deleted
	DeleteByComparisonID(ctx context.Context, comparisonID int64) (int64, error)

	// DeleteByUserID removes all participations for a specific user
	// This is useful for user account deletion
	// Returns the number of participations deleted
	DeleteByUserID(ctx context.Context, userID uuid.UUID) (int64, error)

	// List retrieves a paginated list of all participants
	// page: page number (1-based)
	// pageSize: maximum number of participants to return per page
	// Returns participants ordered by created_at DESC, total count, and error
	List(ctx context.Context, page, pageSize int) ([]*entity.ComparisonParticipant, int64, error)

	// GetByDateRange retrieves participants within a specific date range
	// startDate: start of the date range (inclusive)
	// endDate: end of the date range (inclusive)
	// page: page number (1-based)
	// pageSize: maximum number of participants to return per page
	// Returns participants ordered by created_at DESC, total count, and error
	GetByDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*entity.ComparisonParticipant, int64, error)

	// Count returns the total number of participants in the system
	// This is useful for analytics and administrative dashboards
	Count(ctx context.Context) (int64, error)

	// CountByComparisonID returns the number of participants for a specific comparison
	// This is useful for comparison statistics
	CountByComparisonID(ctx context.Context, comparisonID int64) (int64, error)

	// CountByUserID returns the number of comparisons a user has participated in
	// This is useful for user analytics and gamification
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)

	// Exists checks if a participant exists with the given ID
	// Returns true if the participant exists, false otherwise
	// Returns an error only if there's a system/database error
	Exists(ctx context.Context, id int64) (bool, error)

	// ExistsByComparisonAndUser checks if a user is participating in a specific comparison
	// Returns true if the user is participating, false otherwise
	// Returns an error only if there's a system/database error
	ExistsByComparisonAndUser(ctx context.Context, comparisonID int64, userID uuid.UUID) (bool, error)

	// GetRecentParticipants retrieves recently joined participants
	// hours: number of hours to look back from now
	// page: page number (1-based)
	// pageSize: maximum number of participants to return per page
	// Returns participants ordered by created_at DESC, total count, and error
	GetRecentParticipants(ctx context.Context, hours int, page, pageSize int) ([]*entity.ComparisonParticipant, int64, error)

	// GetActiveUsers retrieves users who have participated in comparisons recently
	// hours: number of hours to look back from now
	// limit: maximum number of users to return
	// Returns user IDs ordered by most recent participation
	GetActiveUsers(ctx context.Context, hours int, limit int) ([]uuid.UUID, error)

	// GetUserParticipationStats retrieves participation statistics for a user
	// Returns total participations, recent participations (last 30 days), and average per month
	GetUserParticipationStats(ctx context.Context, userID uuid.UUID) (total int64, recent int64, avgPerMonth float64, err error)
}
