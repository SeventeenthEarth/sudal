package repo

import (
	"context"

	"github.com/seventeenthearth/sudal/internal/feature/quiz_set/domain/entity"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../../mocks/mock_quiz_set_repository.go -package=mocks -mock_names=QuizSetRepository=MockQuizSetRepository github.com/seventeenthearth/sudal/internal/feature/quiz_set/domain/repo QuizSetRepository

// QuizSetRepository defines the interface for quiz set data access operations
// This interface abstracts the data layer and supports both PostgreSQL and Redis implementations
// following the Repository Pattern to maintain clean separation between domain and data layers.
//
// Implementation Strategy:
// - Write Operations (Create/Update): Write to PostgreSQL first, then update/invalidate Redis cache
// - Read Operations: Attempt Redis first (cache), fallback to PostgreSQL on cache miss
//
// Error Handling:
// - Use domain-specific sentinel errors (entity.ErrQuizSetNotFound, entity.ErrQuizSetAlreadyExists, etc.)
// - Wrap infrastructure errors with appropriate context
// - Ensure consistent error types across different implementations
type QuizSetRepository interface {
	// Create creates a new quiz set in the system
	// The quiz set must have a valid title
	// Returns entity.ErrQuizSetAlreadyExists if a quiz set with the same title already exists
	// Returns entity.ErrQuizSetTitleRequired if the title is empty
	// Returns entity.ErrQuizSetTitleTooLong if the title exceeds maximum length
	Create(ctx context.Context, quizSet *entity.QuizSet) (*entity.QuizSet, error)

	// GetByID retrieves a quiz set by its unique ID
	// Returns entity.ErrQuizSetNotFound if no quiz set exists with the given ID
	// Returns entity.ErrQuizSetInvalidID if the provided ID is invalid
	GetByID(ctx context.Context, id int64) (*entity.QuizSet, error)

	// Update updates an existing quiz set's information
	// The quiz set ID must exist in the system
	// Only non-zero/non-nil fields will be updated (partial updates supported)
	// The UpdatedAt timestamp will be automatically set to the current time
	// Returns entity.ErrQuizSetNotFound if no quiz set exists with the given ID
	Update(ctx context.Context, quizSet *entity.QuizSet) (*entity.QuizSet, error)

	// Delete removes a quiz set from the system (soft delete recommended)
	// This operation should also clean up related data (questions, quiz results, etc.)
	// Returns entity.ErrQuizSetNotFound if no quiz set exists with the given ID
	// Note: Consider implementing soft delete for data retention and audit purposes
	Delete(ctx context.Context, id int64) error

	// List retrieves a paginated list of quiz sets
	// page: page number (1-based)
	// pageSize: maximum number of quiz sets to return per page
	// Returns quiz sets, total count, and error
	// Returns an empty slice if no quiz sets are found (not an error)
	List(ctx context.Context, page, pageSize int) ([]*entity.QuizSet, int64, error)

	// SearchByTitle searches for quiz sets by title pattern
	// titlePattern: pattern to search for in quiz set titles
	// page: page number (1-based)
	// pageSize: maximum number of quiz sets to return per page
	// Returns quiz sets matching the pattern, total count, and error
	SearchByTitle(ctx context.Context, titlePattern string, page, pageSize int) ([]*entity.QuizSet, int64, error)

	// Count returns the total number of quiz sets in the system
	// This is useful for pagination calculations and administrative dashboards
	Count(ctx context.Context) (int64, error)

	// Exists checks if a quiz set exists with the given ID
	// Returns true if the quiz set exists, false otherwise
	// Returns an error only if there's a system/database error
	Exists(ctx context.Context, id int64) (bool, error)
}
