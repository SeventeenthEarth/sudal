package repo

import (
	"context"

	"github.com/seventeenthearth/sudal/internal/feature/quiz/domain/entity"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../../mocks/mock_quiz_repository.go -package=mocks -mock_names=QuizRepository=MockQuizRepository github.com/seventeenthearth/sudal/internal/feature/quiz/domain/repo QuizRepository

// QuizRepository defines the interface for quiz data access operations
// This interface abstracts the data layer and supports both PostgreSQL and Redis implementations
// following the Repository Pattern to maintain clean separation between domain and data layers.
//
// Implementation Strategy:
// - Write Operations (Create/Update): Write to PostgreSQL first, then update/invalidate Redis cache
// - Read Operations: Attempt Redis first (cache), fallback to PostgreSQL on cache miss
//
// Error Handling:
// - Use domain-specific sentinel errors (entity.ErrQuizNotFound, entity.ErrQuizAlreadyExists, etc.)
// - Wrap infrastructure errors with appropriate context
// - Ensure consistent error types across different implementations
type QuizRepository interface {
	// Create creates a new quiz for a quiz set
	// The quiz must have valid text, options, and belong to an existing quiz set
	// Returns entity.ErrQuizInvalidQuizSetID if the quiz set doesn't exist
	// Returns entity.ErrQuizOrderConflict if the order conflicts with existing quizzes
	Create(ctx context.Context, quiz *entity.Quiz) (*entity.Quiz, error)

	// CreateBatch creates multiple quizzes for a quiz set in a single transaction
	// All quizzes must belong to the same quiz set and have unique orders
	// Returns entity.ErrQuizInvalidQuizSetID if the quiz set doesn't exist
	// Returns entity.ErrQuizOrderConflict if any order conflicts with existing quizzes
	CreateBatch(ctx context.Context, quizzes []*entity.Quiz) ([]*entity.Quiz, error)

	// GetByID retrieves a quiz by its unique ID
	// Returns entity.ErrQuizNotFound if no quiz exists with the given ID
	// Returns entity.ErrQuizInvalidID if the provided ID is invalid
	GetByID(ctx context.Context, id int64) (*entity.Quiz, error)

	// GetByQuizSetID retrieves all quizzes for a specific quiz set
	// Quizzes are returned ordered by their quiz_order
	// Returns an empty slice if no quizzes are found (not an error)
	GetByQuizSetID(ctx context.Context, quizSetID int64) ([]*entity.Quiz, error)

	// GetByQuizSetIDPaginated retrieves quizzes for a quiz set with pagination
	// Quizzes are returned ordered by their quiz_order
	// page: page number (1-based)
	// pageSize: maximum number of quizzes to return per page
	// Returns quizzes, total count, and error
	GetByQuizSetIDPaginated(ctx context.Context, quizSetID int64, page, pageSize int) ([]*entity.Quiz, int64, error)

	// Update updates an existing quiz's information
	// The quiz ID must exist in the system
	// Returns entity.ErrQuizNotFound if no quiz exists with the given ID
	// Returns entity.ErrQuizOrderConflict if the new order conflicts with existing quizzes
	Update(ctx context.Context, quiz *entity.Quiz) (*entity.Quiz, error)

	// Delete removes a quiz from the system
	// This operation should also update the quiz set's quiz count
	// Returns entity.ErrQuizNotFound if no quiz exists with the given ID
	Delete(ctx context.Context, id int64) error

	// DeleteByQuizSetID removes all quizzes for a specific quiz set
	// This is useful when deleting an entire quiz set
	// Returns the number of quizzes deleted
	DeleteByQuizSetID(ctx context.Context, quizSetID int64) (int64, error)

	// ReorderQuizzes updates the order of multiple quizzes in a quiz set
	// quizIDs: slice of quiz IDs in the desired order
	// The first ID will have order 1, second will have order 2, etc.
	// Returns entity.ErrQuizNotFound if any quiz ID doesn't exist
	ReorderQuizzes(ctx context.Context, quizSetID int64, quizIDs []int64) error

	// Count returns the total number of quizzes in the system
	// This is useful for analytics and administrative dashboards
	Count(ctx context.Context) (int64, error)

	// CountByQuizSetID returns the number of quizzes for a specific quiz set
	// This is useful for quiz set statistics and validation
	CountByQuizSetID(ctx context.Context, quizSetID int64) (int64, error)

	// Exists checks if a quiz exists with the given ID
	// Returns true if the quiz exists, false otherwise
	// Returns an error only if there's a system/database error
	Exists(ctx context.Context, id int64) (bool, error)

	// GetMaxOrderByQuizSetID returns the highest quiz order for a quiz set
	// This is useful when adding new quizzes to determine the next order
	// Returns 0 if no quizzes exist for the quiz set
	GetMaxOrderByQuizSetID(ctx context.Context, quizSetID int64) (int, error)

	// SearchByText searches for quizzes by text content
	// textPattern: pattern to search for in quiz text
	// page: page number (1-based)
	// pageSize: maximum number of quizzes to return per page
	// Returns quizzes matching the pattern, total count, and error
	SearchByText(ctx context.Context, textPattern string, page, pageSize int) ([]*entity.Quiz, int64, error)
}
