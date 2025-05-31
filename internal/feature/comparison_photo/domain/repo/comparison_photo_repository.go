package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/seventeenthearth/sudal/internal/feature/comparison_photo/domain/entity"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../../mocks/mock_comparison_photo_repository.go -package=mocks -mock_names=ComparisonPhotoRepository=MockComparisonPhotoRepository github.com/seventeenthearth/sudal/internal/feature/comparison_photo/domain/repo ComparisonPhotoRepository

// ComparisonPhotoRepository defines the protocol for comparison photo data access operations
// This protocol abstracts the data layer and supports both PostgreSQL and Redis implementations
// following the Repository Pattern to maintain clean separation between domain and data layers.
//
// Implementation Strategy:
// - Write Operations (Create/Delete): Write to PostgreSQL first, then update/invalidate Redis cache
// - Read Operations: Attempt Redis first (cache), fallback to PostgreSQL on cache miss
//
// Error Handling:
// - Use domain-specific sentinel errors (entity.ErrComparisonPhotoNotFound, entity.ErrComparisonPhotoPermissionDenied, etc.)
// - Wrap infrastructure errors with appropriate context
// - Ensure consistent error types across different implementations
type ComparisonPhotoRepository interface {
	// Create adds a photo to a comparison event
	// The photo must have a valid comparison ID and uploader user ID
	// Returns entity.ErrComparisonPhotoInvalidComparisonID if the comparison doesn't exist
	// Returns entity.ErrComparisonPhotoUploadFailed if the photo upload fails
	Create(ctx context.Context, photo *entity.ComparisonPhoto) (*entity.ComparisonPhoto, error)

	// GetByID retrieves a specific photo by its ID
	// Returns entity.ErrComparisonPhotoNotFound if the photo doesn't exist
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ComparisonPhoto, error)

	// GetByComparisonID retrieves all photos for a specific comparison
	// Returns photos ordered by uploaded_at ASC (chronological order)
	// Returns an empty slice if no photos are found (not an error)
	GetByComparisonID(ctx context.Context, comparisonID int64) ([]*entity.ComparisonPhoto, error)

	// GetByUploaderUserID retrieves all photos uploaded by a specific user
	// page: page number (1-based)
	// pageSize: maximum number of photos to return per page
	// Returns photos ordered by uploaded_at DESC, total count, and error
	GetByUploaderUserID(ctx context.Context, uploaderUserID uuid.UUID, page, pageSize int) ([]*entity.ComparisonPhoto, int64, error)

	// Update updates an existing photo's information (mainly photo URL)
	// Returns entity.ErrComparisonPhotoNotFound if the photo doesn't exist
	Update(ctx context.Context, photo *entity.ComparisonPhoto) (*entity.ComparisonPhoto, error)

	// Delete removes a photo from a comparison event
	// Only the uploader or comparison owner should be able to remove photos
	// Returns entity.ErrComparisonPhotoNotFound if the photo doesn't exist
	// Returns entity.ErrComparisonPhotoPermissionDenied if the user doesn't have permission
	Delete(ctx context.Context, id uuid.UUID, requestingUserID uuid.UUID) error

	// DeleteByComparisonID removes all photos for a specific comparison
	// This is useful when deleting an entire comparison
	// Returns the number of photos deleted
	DeleteByComparisonID(ctx context.Context, comparisonID int64) (int64, error)

	// DeleteByUploaderUserID removes all photos uploaded by a specific user
	// This is useful for user account deletion
	// Returns the number of photos deleted
	DeleteByUploaderUserID(ctx context.Context, uploaderUserID uuid.UUID) (int64, error)

	// List retrieves a paginated list of all photos
	// page: page number (1-based)
	// pageSize: maximum number of photos to return per page
	// Returns photos ordered by uploaded_at DESC, total count, and error
	List(ctx context.Context, page, pageSize int) ([]*entity.ComparisonPhoto, int64, error)

	// GetByDateRange retrieves photos within a specific date range
	// startDate: start of the date range (inclusive)
	// endDate: end of the date range (inclusive)
	// page: page number (1-based)
	// pageSize: maximum number of photos to return per page
	// Returns photos ordered by uploaded_at DESC, total count, and error
	GetByDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*entity.ComparisonPhoto, int64, error)

	// Count returns the total number of photos in the system
	// This is useful for analytics and storage management
	Count(ctx context.Context) (int64, error)

	// CountByComparisonID returns the number of photos for a specific comparison
	// This is useful for comparison statistics
	CountByComparisonID(ctx context.Context, comparisonID int64) (int64, error)

	// CountByUploaderUserID returns the number of photos uploaded by a specific user
	// This is useful for user analytics and storage quotas
	CountByUploaderUserID(ctx context.Context, uploaderUserID uuid.UUID) (int64, error)

	// Exists checks if a photo exists with the given ID
	// Returns true if the photo exists, false otherwise
	// Returns an error only if there's a system/database error
	Exists(ctx context.Context, id uuid.UUID) (bool, error)

	// GetRecentPhotos retrieves recently uploaded photos
	// hours: number of hours to look back from now
	// page: page number (1-based)
	// pageSize: maximum number of photos to return per page
	// Returns photos ordered by uploaded_at DESC, total count, and error
	GetRecentPhotos(ctx context.Context, hours int, page, pageSize int) ([]*entity.ComparisonPhoto, int64, error)

	// GetPhotosByFileExtension retrieves photos by file extension
	// extension: file extension to filter by (e.g., ".jpg", ".png")
	// page: page number (1-based)
	// pageSize: maximum number of photos to return per page
	// Returns photos ordered by uploaded_at DESC, total count, and error
	GetPhotosByFileExtension(ctx context.Context, extension string, page, pageSize int) ([]*entity.ComparisonPhoto, int64, error)

	// GetActiveUploaders retrieves users who have uploaded photos recently
	// hours: number of hours to look back from now
	// limit: maximum number of users to return
	// Returns user IDs ordered by most recent upload
	GetActiveUploaders(ctx context.Context, hours int, limit int) ([]uuid.UUID, error)

	// GetUserUploadStats retrieves upload statistics for a user
	// Returns total uploads, recent uploads (last 30 days), and average per month
	GetUserUploadStats(ctx context.Context, userID uuid.UUID) (total int64, recent int64, avgPerMonth float64, err error)

	// ValidateUserCanUpload checks if a user can upload a photo to a specific comparison
	// This includes checking permissions, comparison status, etc.
	// Returns true if the user can upload, false otherwise
	// Returns an error only if there's a system/database error
	ValidateUserCanUpload(ctx context.Context, comparisonID int64, userID uuid.UUID) (bool, error)

	// ValidateUserCanDelete checks if a user can delete a specific photo
	// This includes checking if the user is the uploader or has admin permissions
	// Returns true if the user can delete, false otherwise
	// Returns an error only if there's a system/database error
	ValidateUserCanDelete(ctx context.Context, photoID uuid.UUID, userID uuid.UUID) (bool, error)
}
