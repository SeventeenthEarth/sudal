package repo

import (
	"context"
	"time"

	"github.com/seventeenthearth/sudal/internal/feature/comparison/domain/entity"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../../mocks/mock_comparison_repository.go -package=mocks -mock_names=ComparisonRepository=MockComparisonRepository github.com/seventeenthearth/sudal/internal/feature/comparison/domain/repo ComparisonRepository

// ComparisonRepository defines the interface for comparison data access operations
// This interface abstracts the data layer and supports both PostgreSQL and Redis implementations
// following the Repository Pattern to maintain clean separation between domain and data layers.
//
// Implementation Strategy:
// - Write Operations (Create/Update): Write to PostgreSQL first, then update/invalidate Redis cache
// - Read Operations: Attempt Redis first (cache), fallback to PostgreSQL on cache miss
//
// Error Handling:
// - Use domain-specific sentinel errors (entity.ErrComparisonNotFound, entity.ErrComparisonAlreadyExists, etc.)
// - Wrap infrastructure errors with appropriate context
// - Ensure consistent error types across different implementations
type ComparisonRepository interface {
	// Create creates a new comparison event
	// The comparison must have a valid quiz set ID and room ID
	// Returns entity.ErrComparisonAlreadyExists if a comparison with the same room ID already exists
	// Returns entity.ErrComparisonInvalidQuizSetID if the quiz set doesn't exist
	// Returns entity.ErrComparisonRoomIDRequired if the room ID is empty
	Create(ctx context.Context, comparison *entity.Comparison) (*entity.Comparison, error)

	// GetByID retrieves a comparison by its unique ID
	// Returns entity.ErrComparisonNotFound if no comparison exists with the given ID
	// Returns entity.ErrComparisonInvalidID if the provided ID is invalid
	GetByID(ctx context.Context, id int64) (*entity.Comparison, error)

	// GetByRoomID retrieves a comparison by its room ID
	// Returns entity.ErrComparisonNotFound if no comparison exists with the given room ID
	GetByRoomID(ctx context.Context, roomID string) (*entity.Comparison, error)

	// Update updates an existing comparison's information
	// The comparison ID must exist in the system
	// Returns entity.ErrComparisonNotFound if no comparison exists with the given ID
	Update(ctx context.Context, comparison *entity.Comparison) (*entity.Comparison, error)

	// Delete removes a comparison from the system
	// This operation should also clean up related data (participants, photos, etc.)
	// Returns entity.ErrComparisonNotFound if no comparison exists with the given ID
	Delete(ctx context.Context, id int64) error

	// List retrieves a paginated list of comparisons
	// page: page number (1-based)
	// pageSize: maximum number of comparisons to return per page
	// Returns comparisons ordered by created_at DESC, total count, and error
	List(ctx context.Context, page, pageSize int) ([]*entity.Comparison, int64, error)

	// GetByQuizSetID retrieves all comparisons for a specific quiz set
	// page: page number (1-based)
	// pageSize: maximum number of comparisons to return per page
	// Returns comparisons ordered by created_at DESC, total count, and error
	GetByQuizSetID(ctx context.Context, quizSetID int64, page, pageSize int) ([]*entity.Comparison, int64, error)

	// GetByDateRange retrieves comparisons within a specific date range
	// startDate: start of the date range (inclusive)
	// endDate: end of the date range (inclusive)
	// page: page number (1-based)
	// pageSize: maximum number of comparisons to return per page
	// Returns comparisons ordered by created_at DESC, total count, and error
	GetByDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*entity.Comparison, int64, error)

	// GetActiveComparisons retrieves comparisons that have recent activity
	// hoursThreshold: number of hours to consider as "recent activity"
	// page: page number (1-based)
	// pageSize: maximum number of comparisons to return per page
	// Returns comparisons ordered by created_at DESC, total count, and error
	GetActiveComparisons(ctx context.Context, hoursThreshold int, page, pageSize int) ([]*entity.Comparison, int64, error)

	// SearchByRoomID searches for comparisons by room ID pattern
	// roomIDPattern: pattern to search for in room IDs
	// page: page number (1-based)
	// pageSize: maximum number of comparisons to return per page
	// Returns comparisons matching the pattern, total count, and error
	SearchByRoomID(ctx context.Context, roomIDPattern string, page, pageSize int) ([]*entity.Comparison, int64, error)

	// Count returns the total number of comparisons in the system
	// This is useful for analytics and administrative dashboards
	Count(ctx context.Context) (int64, error)

	// CountByQuizSetID returns the number of comparisons for a specific quiz set
	// This is useful for quiz set analytics and popularity metrics
	CountByQuizSetID(ctx context.Context, quizSetID int64) (int64, error)

	// Exists checks if a comparison exists with the given ID
	// Returns true if the comparison exists, false otherwise
	// Returns an error only if there's a system/database error
	Exists(ctx context.Context, id int64) (bool, error)

	// ExistsByRoomID checks if a comparison exists with the given room ID
	// Returns true if the comparison exists, false otherwise
	// Returns an error only if there's a system/database error
	ExistsByRoomID(ctx context.Context, roomID string) (bool, error)

	// GetRecentComparisons retrieves recently created comparisons
	// hours: number of hours to look back from now
	// page: page number (1-based)
	// pageSize: maximum number of comparisons to return per page
	// Returns comparisons ordered by created_at DESC, total count, and error
	GetRecentComparisons(ctx context.Context, hours int, page, pageSize int) ([]*entity.Comparison, int64, error)
}
