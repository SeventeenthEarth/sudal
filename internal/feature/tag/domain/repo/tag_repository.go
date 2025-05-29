package repo

import (
	"context"

	"github.com/seventeenthearth/sudal/internal/feature/tag/domain/entity"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../../mocks/mock_tag_repository.go -package=mocks -mock_names=TagRepository=MockTagRepository github.com/seventeenthearth/sudal/internal/feature/tag/domain/repo TagRepository

// TagRepository defines the interface for tag data access operations
// This interface abstracts the data layer and supports both PostgreSQL and Redis implementations
// following the Repository Pattern to maintain clean separation between domain and data layers.
//
// Implementation Strategy:
// - Write Operations (Create/Update): Write to PostgreSQL first, then update/invalidate Redis cache
// - Read Operations: Attempt Redis first (cache), fallback to PostgreSQL on cache miss
//
// Error Handling:
// - Use domain-specific sentinel errors (entity.ErrTagNotFound, entity.ErrTagAlreadyExists, etc.)
// - Wrap infrastructure errors with appropriate context
// - Ensure consistent error types across different implementations
type TagRepository interface {
	// Create creates a new tag in the system
	// The tag must have a unique name
	// Returns entity.ErrTagAlreadyExists if a tag with the same name already exists
	// Returns entity.ErrTagNameRequired if the name is empty
	Create(ctx context.Context, tag *entity.Tag) (*entity.Tag, error)

	// GetByID retrieves a tag by its unique ID
	// Returns entity.ErrTagNotFound if no tag exists with the given ID
	// Returns entity.ErrTagInvalidID if the provided ID is invalid
	GetByID(ctx context.Context, id int64) (*entity.Tag, error)

	// GetByName retrieves a tag by its name
	// Returns entity.ErrTagNotFound if no tag exists with the given name
	GetByName(ctx context.Context, name string) (*entity.Tag, error)

	// GetByNames retrieves multiple tags by their names
	// Returns only the tags that exist, missing tags are not included in the result
	// Returns an empty slice if none of the tags exist (not an error)
	GetByNames(ctx context.Context, names []string) ([]*entity.Tag, error)

	// Update updates an existing tag's information
	// Returns entity.ErrTagNotFound if no tag exists with the given ID
	// Returns entity.ErrTagAlreadyExists if the new name conflicts with an existing tag
	Update(ctx context.Context, tag *entity.Tag) (*entity.Tag, error)

	// Delete removes a tag from the system
	// This operation should check if the tag is still in use before deletion
	// Returns entity.ErrTagNotFound if no tag exists with the given ID
	// Returns entity.ErrTagInUse if the tag is still associated with quiz sets
	Delete(ctx context.Context, id int64) error

	// List retrieves a paginated list of all tags
	// page: page number (1-based)
	// pageSize: maximum number of tags to return per page
	// Returns tags ordered by name ASC, total count, and error
	List(ctx context.Context, page, pageSize int) ([]*entity.Tag, int64, error)

	// SearchByName searches for tags by name pattern
	// namePattern: pattern to search for in tag names (case-insensitive)
	// page: page number (1-based)
	// pageSize: maximum number of tags to return per page
	// Returns tags matching the pattern ordered by name ASC, total count, and error
	SearchByName(ctx context.Context, namePattern string, page, pageSize int) ([]*entity.Tag, int64, error)

	// GetPopularTags retrieves the most popular tags based on usage
	// This requires counting associations with quiz sets
	// limit: maximum number of tags to return
	// Returns tags ordered by usage count DESC
	GetPopularTags(ctx context.Context, limit int) ([]*entity.Tag, error)

	// GetRecentTags retrieves recently created tags
	// limit: maximum number of tags to return
	// Returns tags ordered by created_at DESC
	GetRecentTags(ctx context.Context, limit int) ([]*entity.Tag, error)

	// Count returns the total number of tags in the system
	// This is useful for pagination calculations and administrative dashboards
	Count(ctx context.Context) (int64, error)

	// Exists checks if a tag exists with the given ID
	// Returns true if the tag exists, false otherwise
	// Returns an error only if there's a system/database error
	Exists(ctx context.Context, id int64) (bool, error)

	// ExistsByName checks if a tag exists with the given name
	// Returns true if the tag exists, false otherwise
	// Returns an error only if there's a system/database error
	ExistsByName(ctx context.Context, name string) (bool, error)

	// GetUsageCount returns the number of quiz sets using a specific tag
	// This is useful for determining tag popularity and preventing deletion of used tags
	GetUsageCount(ctx context.Context, tagID int64) (int64, error)

	// GetUnusedTags retrieves tags that are not associated with any quiz sets
	// This is useful for cleanup operations
	// page: page number (1-based)
	// pageSize: maximum number of tags to return per page
	// Returns unused tags ordered by created_at ASC, total count, and error
	GetUnusedTags(ctx context.Context, page, pageSize int) ([]*entity.Tag, int64, error)

	// BulkCreate creates multiple tags in a single transaction
	// Skips tags that already exist (by name) and returns only the newly created ones
	// Returns the created tags and any error that occurred
	BulkCreate(ctx context.Context, tags []*entity.Tag) ([]*entity.Tag, error)

	// BulkDelete removes multiple tags from the system
	// Only deletes tags that are not in use
	// Returns the number of tags successfully deleted and any error
	BulkDelete(ctx context.Context, tagIDs []int64) (int64, error)
}
