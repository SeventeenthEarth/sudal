package entity

import (
	"time"

	"github.com/google/uuid"
)

// ComparisonPhoto represents a photo uploaded for a comparison event
// This domain model encapsulates all photo-related data and business rules
type ComparisonPhoto struct {
	// ID is the unique identifier for the photo (UUID for security)
	ID uuid.UUID `json:"id"`

	// ComparisonID is the foreign key reference to the comparison event
	ComparisonID int64 `json:"comparison_id"`

	// UploaderUserID is the foreign key reference to the user who uploaded the photo
	UploaderUserID uuid.UUID `json:"uploader_user_id"`

	// PhotoURL is the URL to the uploaded photo
	PhotoURL string `json:"photo_url"`

	// UploadedAt is the timestamp when the photo was uploaded (UTC)
	UploadedAt time.Time `json:"uploaded_at"`
}

// NewComparisonPhoto creates a new ComparisonPhoto with the provided parameters
// This constructor ensures required fields are set and provides sensible defaults
func NewComparisonPhoto(comparisonID int64, uploaderUserID uuid.UUID, photoURL string) *ComparisonPhoto {
	return &ComparisonPhoto{
		ID:             uuid.New(),
		ComparisonID:   comparisonID,
		UploaderUserID: uploaderUserID,
		PhotoURL:       photoURL,
		UploadedAt:     time.Now().UTC(),
	}
}

// UpdatePhotoURL updates the photo URL
// This might be used if the photo is moved to a different storage location
func (cp *ComparisonPhoto) UpdatePhotoURL(photoURL string) {
	cp.PhotoURL = photoURL
}

// IsUploadedBy checks if the photo was uploaded by the specified user
func (cp *ComparisonPhoto) IsUploadedBy(userID uuid.UUID) bool {
	return cp.UploaderUserID == userID
}

// GetFileExtension extracts the file extension from the photo URL
// This is useful for determining the file type
func (cp *ComparisonPhoto) GetFileExtension() string {
	url := cp.PhotoURL
	for i := len(url) - 1; i >= 0; i-- {
		if url[i] == '.' {
			return url[i:]
		}
		if url[i] == '/' || url[i] == '?' {
			break
		}
	}
	return ""
}

// IsImageFile checks if the photo URL appears to be an image file
// This is a basic check based on common image file extensions
func (cp *ComparisonPhoto) IsImageFile() bool {
	ext := cp.GetFileExtension()
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg":
		return true
	default:
		return false
	}
}

// GetUploadAge returns the duration since the photo was uploaded
func (cp *ComparisonPhoto) GetUploadAge() time.Duration {
	return time.Since(cp.UploadedAt)
}

// IsRecentUpload checks if the photo was uploaded recently (within the last hour)
func (cp *ComparisonPhoto) IsRecentUpload() bool {
	return cp.GetUploadAge() < time.Hour
}

// Validate performs basic validation on the comparison photo
func (cp *ComparisonPhoto) Validate() error {
	if cp.ID == uuid.Nil {
		return ErrComparisonPhotoInvalidID
	}
	if cp.ComparisonID <= 0 {
		return ErrComparisonPhotoInvalidComparisonID
	}
	if cp.UploaderUserID == uuid.Nil {
		return ErrComparisonPhotoInvalidUploaderUserID
	}
	if len(cp.PhotoURL) == 0 {
		return ErrComparisonPhotoURLRequired
	}
	if len(cp.PhotoURL) > 2048 { // Reasonable URL length limit
		return ErrComparisonPhotoURLTooLong
	}
	if cp.UploadedAt.IsZero() {
		return ErrComparisonPhotoInvalidUploadedTime
	}
	return nil
}
