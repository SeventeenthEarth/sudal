package repo

import (
	"context"
	"time"

	"github.com/seventeenthearth/sudal/internal/feature/quiz_set_tag/domain/entity"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../../mocks/mock_quiz_set_tag_repository.go -package=mocks -mock_names=QuizSetTagRepository=MockQuizSetTagRepository github.com/seventeenthearth/sudal/internal/feature/quiz_set_tag/domain/repo QuizSetTagRepository

// QuizSetTagRepository defines the interface for quiz set tag association data access operations
// This interface abstracts the data layer and supports both PostgreSQL and Redis implementations
// following the Repository Pattern to maintain clean separation between domain and data layers.
//
// Implementation Strategy:
// - Write Operations (Create/Delete): Write to PostgreSQL first, then update/invalidate Redis cache
// - Read Operations: Attempt Redis first (cache), fallback to PostgreSQL on cache miss
//
// Error Handling:
// - Use domain-specific sentinel errors (entity.ErrQuizSetTagNotFound, entity.ErrQuizSetTagAlreadyExists, etc.)
// - Wrap infrastructure errors with appropriate context
// - Ensure consistent error types across different implementations
type QuizSetTagRepository interface {
	// Create creates a new association between a quiz set and a tag
	// Returns entity.ErrQuizSetTagAlreadyExists if the association already exists
	// Returns entity.ErrQuizSetTagInvalidQuizSetID if the quiz set doesn't exist
	// Returns entity.ErrQuizSetTagInvalidTagID if the tag doesn't exist
	Create(ctx context.Context, quizSetTag *entity.QuizSetTag) (*entity.QuizSetTag, error)

	// Delete removes an association between a quiz set and a tag
	// Returns entity.ErrQuizSetTagNotFound if the association doesn't exist
	Delete(ctx context.Context, quizSetID, tagID int64) error

	// GetByQuizSetID retrieves all tag associations for a specific quiz set
	// Returns detailed information including tag names, descriptions, and colors
	// Returns an empty slice if no associations are found (not an error)
	GetByQuizSetID(ctx context.Context, quizSetID int64) ([]*entity.QuizSetTagDetail, error)

	// GetByTagID retrieves all quiz set associations for a specific tag
	// page: page number (1-based)
	// pageSize: maximum number of associations to return per page
	// Returns detailed information including quiz set titles
	// Returns associations ordered by assigned_at DESC, total count, and error
	GetByTagID(ctx context.Context, tagID int64, page, pageSize int) ([]*entity.QuizSetTagDetail, int64, error)

	// GetByQuizSetAndTag retrieves a specific association between a quiz set and tag
	// Returns entity.ErrQuizSetTagNotFound if the association doesn't exist
	GetByQuizSetAndTag(ctx context.Context, quizSetID, tagID int64) (*entity.QuizSetTagDetail, error)

	// Exists checks if an association exists between a quiz set and tag
	// Returns true if the association exists, false otherwise
	// Returns an error only if there's a system/database error
	Exists(ctx context.Context, quizSetID, tagID int64) (bool, error)

	// GetTagIDsByQuizSetID retrieves only the tag IDs for a specific quiz set
	// This is useful when you only need the IDs without additional tag details
	// Returns an empty slice if no associations are found (not an error)
	GetTagIDsByQuizSetID(ctx context.Context, quizSetID int64) ([]int64, error)

	// GetQuizSetIDsByTagID retrieves only the quiz set IDs for a specific tag
	// page: page number (1-based)
	// pageSize: maximum number of quiz set IDs to return per page
	// Returns quiz set IDs ordered by assigned_at DESC, total count, and error
	GetQuizSetIDsByTagID(ctx context.Context, tagID int64, page, pageSize int) ([]int64, int64, error)

	// BulkCreate creates multiple associations in a single transaction
	// Skips associations that already exist and returns only the newly created ones
	// Returns the created associations and any error that occurred
	BulkCreate(ctx context.Context, quizSetTags []*entity.QuizSetTag) ([]*entity.QuizSetTag, error)

	// BulkDelete removes multiple associations in a single transaction
	// Returns the number of associations successfully deleted and any error
	BulkDelete(ctx context.Context, associations []entity.QuizSetTagAssociation) (int64, error)

	// ReplaceTagsForQuizSet replaces all tags for a quiz set with a new set of tags
	// This operation removes existing associations and creates new ones in a single transaction
	// Returns the new associations and any error that occurred
	ReplaceTagsForQuizSet(ctx context.Context, quizSetID int64, tagIDs []int64) ([]*entity.QuizSetTag, error)

	// GetRecentAssociations retrieves recently created associations
	// hours: number of hours to look back from now
	// page: page number (1-based)
	// pageSize: maximum number of associations to return per page
	// Returns associations ordered by assigned_at DESC, total count, and error
	GetRecentAssociations(ctx context.Context, hours int, page, pageSize int) ([]*entity.QuizSetTagDetail, int64, error)

	// GetAssociationsByDateRange retrieves associations within a specific date range
	// startDate: start of the date range (inclusive)
	// endDate: end of the date range (inclusive)
	// page: page number (1-based)
	// pageSize: maximum number of associations to return per page
	// Returns associations ordered by assigned_at DESC, total count, and error
	GetAssociationsByDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*entity.QuizSetTagDetail, int64, error)

	// Count returns the total number of quiz set tag associations in the system
	// This is useful for analytics and administrative dashboards
	Count(ctx context.Context) (int64, error)

	// CountByQuizSetID returns the number of tags associated with a specific quiz set
	// This is useful for quiz set statistics and validation
	CountByQuizSetID(ctx context.Context, quizSetID int64) (int64, error)

	// CountByTagID returns the number of quiz sets associated with a specific tag
	// This is useful for tag popularity metrics
	CountByTagID(ctx context.Context, tagID int64) (int64, error)

	// GetMostUsedTags retrieves tags ordered by their usage count (number of quiz sets)
	// limit: maximum number of tags to return
	// Returns tag IDs and their usage counts ordered by count DESC
	GetMostUsedTags(ctx context.Context, limit int) ([]int64, []int64, error)

	// GetLeastUsedTags retrieves tags ordered by their usage count (ascending)
	// This is useful for identifying tags that might need promotion or cleanup
	// limit: maximum number of tags to return
	// Returns tag IDs and their usage counts ordered by count ASC
	GetLeastUsedTags(ctx context.Context, limit int) ([]int64, []int64, error)

	// DeleteByQuizSetID removes all tag associations for a specific quiz set
	// This is useful when deleting an entire quiz set
	// Returns the number of associations deleted
	DeleteByQuizSetID(ctx context.Context, quizSetID int64) (int64, error)

	// DeleteByTagID removes all quiz set associations for a specific tag
	// This is useful when deleting a tag
	// Returns the number of associations deleted
	DeleteByTagID(ctx context.Context, tagID int64) (int64, error)
}
