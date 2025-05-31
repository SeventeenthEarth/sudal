package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/seventeenthearth/sudal/internal/feature/quiz_result/domain/entity"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../../mocks/mock_quiz_result_repository.go -package=mocks -mock_names=QuizResultRepository=MockQuizResultRepository github.com/seventeenthearth/sudal/internal/feature/quiz_result/domain/repo QuizResultRepository

// QuizResultRepository defines the protocol for quiz result data access operations
// This protocol abstracts the data layer and supports both PostgreSQL and Redis implementations
// following the Repository Pattern to maintain clean separation between domain and data layers.
//
// Implementation Strategy:
// - Write Operations (Create/Update): Write to PostgreSQL first, then update/invalidate Redis cache
// - Read Operations: Attempt Redis first (cache), fallback to PostgreSQL on cache miss
//
// Error Handling:
// - Use domain-specific sentinel errors (entity.ErrQuizResultNotFound, entity.ErrQuizResultAlreadyExists, etc.)
// - Wrap infrastructure errors with appropriate context
// - Ensure consistent error types across different implementations
type QuizResultRepository interface {
	// Create creates a new quiz result record
	// The quiz result must have valid user ID, quiz set ID, and answers
	// Returns entity.ErrQuizResultInvalidUserID if the user ID is invalid
	// Returns entity.ErrQuizResultInvalidQuizSetID if the quiz set doesn't exist
	// Returns entity.ErrQuizResultNoAnswers if no answers are provided
	Create(ctx context.Context, result *entity.QuizResult) (*entity.QuizResult, error)

	// GetByID retrieves a quiz result by its unique ID
	// Returns entity.ErrQuizResultNotFound if no quiz result exists with the given ID
	// Returns entity.ErrQuizResultInvalidID if the provided ID is invalid
	GetByID(ctx context.Context, id int64) (*entity.QuizResult, error)

	// GetByUserID retrieves all quiz results for a specific user
	// page: page number (1-based)
	// pageSize: maximum number of results to return per page
	// Returns quiz results ordered by submitted_at DESC, total count, and error
	// Returns an empty slice if no quiz results are found (not an error)
	GetByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*entity.QuizResult, int64, error)

	// GetByQuizSetID retrieves all quiz results for a specific quiz set
	// page: page number (1-based)
	// pageSize: maximum number of results to return per page
	// Returns quiz results ordered by submitted_at DESC, total count, and error
	GetByQuizSetID(ctx context.Context, quizSetID int64, page, pageSize int) ([]*entity.QuizResult, int64, error)

	// GetByUserAndQuizSet retrieves quiz results for a specific user and quiz set
	// page: page number (1-based)
	// pageSize: maximum number of results to return per page
	// Returns quiz results ordered by submitted_at DESC, total count, and error
	GetByUserAndQuizSet(ctx context.Context, userID uuid.UUID, quizSetID int64, page, pageSize int) ([]*entity.QuizResult, int64, error)

	// GetLatestByUserAndQuizSet retrieves the most recent quiz result for a user and quiz set
	// This is useful for comparison features where we need the user's latest attempt
	// Returns entity.ErrQuizResultNotFound if no quiz result exists for the user and quiz set
	GetLatestByUserAndQuizSet(ctx context.Context, userID uuid.UUID, quizSetID int64) (*entity.QuizResult, error)

	// GetByDateRange retrieves quiz results within a specific date range
	// userID: optional user filter (uuid.Nil means all users)
	// quizSetID: optional quiz set filter (0 means all quiz sets)
	// startDate: start of the date range (inclusive)
	// endDate: end of the date range (inclusive)
	// page: page number (1-based)
	// pageSize: maximum number of results to return per page
	// Returns quiz results ordered by submitted_at DESC, total count, and error
	GetByDateRange(ctx context.Context, userID uuid.UUID, quizSetID int64, startDate, endDate time.Time, page, pageSize int) ([]*entity.QuizResult, int64, error)

	// Update updates an existing quiz result's information
	// Note: Updating quiz results should be rare and carefully controlled
	// Returns entity.ErrQuizResultNotFound if no quiz result exists with the given ID
	Update(ctx context.Context, result *entity.QuizResult) (*entity.QuizResult, error)

	// Delete removes a quiz result from the system
	// Note: Deleting quiz results should be rare and carefully controlled
	// Returns entity.ErrQuizResultNotFound if no quiz result exists with the given ID
	Delete(ctx context.Context, id int64) error

	// DeleteByUserID removes all quiz results for a specific user
	// This is useful for user account deletion
	// Returns the number of quiz results deleted
	DeleteByUserID(ctx context.Context, userID uuid.UUID) (int64, error)

	// DeleteByQuizSetID removes all quiz results for a specific quiz set
	// This is useful when deleting an entire quiz set
	// Returns the number of quiz results deleted
	DeleteByQuizSetID(ctx context.Context, quizSetID int64) (int64, error)

	// Count returns the total number of quiz results in the system
	// This is useful for analytics and administrative dashboards
	Count(ctx context.Context) (int64, error)

	// CountByUserID returns the number of quiz results for a specific user
	// This is useful for user analytics and gamification
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)

	// CountByQuizSetID returns the number of quiz results for a specific quiz set
	// This is useful for quiz set analytics and popularity metrics
	CountByQuizSetID(ctx context.Context, quizSetID int64) (int64, error)

	// Exists checks if a quiz result exists with the given ID
	// Returns true if the quiz result exists, false otherwise
	// Returns an error only if there's a system/database error
	Exists(ctx context.Context, id int64) (bool, error)

	// ExistsByUserAndQuizSet checks if a user has taken a specific quiz set
	// Returns true if the user has at least one quiz result for the quiz set
	// Returns an error only if there's a system/database error
	ExistsByUserAndQuizSet(ctx context.Context, userID uuid.UUID, quizSetID int64) (bool, error)

	// GetRecentResults retrieves recently submitted quiz results
	// hours: number of hours to look back from now
	// page: page number (1-based)
	// pageSize: maximum number of results to return per page
	// Returns quiz results ordered by submitted_at DESC, total count, and error
	GetRecentResults(ctx context.Context, hours int, page, pageSize int) ([]*entity.QuizResult, int64, error)

	// GetTopScorers retrieves users with the highest scores for a specific quiz set
	// This requires calculating scores based on correct answers (implementation dependent)
	// quizSetID: the quiz set to get top scorers for
	// limit: maximum number of top scorers to return
	// Returns user IDs and their scores ordered by score DESC
	GetTopScorers(ctx context.Context, quizSetID int64, limit int) ([]uuid.UUID, []int, error)
}
