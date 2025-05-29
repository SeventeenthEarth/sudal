package entity

import "errors"

// Comparison Photo Errors
var (
	// ErrComparisonPhotoNotFound is returned when a comparison photo cannot be found
	ErrComparisonPhotoNotFound = errors.New("comparison photo not found")

	// ErrComparisonPhotoAlreadyExists is returned when attempting to create a photo that already exists
	ErrComparisonPhotoAlreadyExists = errors.New("comparison photo already exists")

	// ErrComparisonPhotoInvalidID is returned when a photo ID is invalid
	ErrComparisonPhotoInvalidID = errors.New("invalid comparison photo ID")

	// ErrComparisonPhotoInvalidComparisonID is returned when the comparison ID is invalid
	ErrComparisonPhotoInvalidComparisonID = errors.New("invalid comparison ID for photo")

	// ErrComparisonPhotoInvalidUploaderUserID is returned when the uploader user ID is invalid
	ErrComparisonPhotoInvalidUploaderUserID = errors.New("invalid uploader user ID for comparison photo")

	// ErrComparisonPhotoURLRequired is returned when a photo URL is empty
	ErrComparisonPhotoURLRequired = errors.New("comparison photo URL is required")

	// ErrComparisonPhotoURLTooLong is returned when a photo URL exceeds the maximum length
	ErrComparisonPhotoURLTooLong = errors.New("comparison photo URL is too long (maximum 2048 characters)")

	// ErrComparisonPhotoInvalidUploadedTime is returned when the uploaded time is invalid
	ErrComparisonPhotoInvalidUploadedTime = errors.New("invalid comparison photo uploaded time")

	// ErrComparisonPhotoUploadFailed is returned when photo upload fails
	ErrComparisonPhotoUploadFailed = errors.New("comparison photo upload failed")

	// ErrComparisonPhotoDeleteFailed is returned when photo deletion fails
	ErrComparisonPhotoDeleteFailed = errors.New("comparison photo deletion failed")

	// ErrComparisonPhotoPermissionDenied is returned when a user doesn't have permission to access/modify a photo
	ErrComparisonPhotoPermissionDenied = errors.New("permission denied for comparison photo operation")

	// ErrComparisonPhotoInvalidPagination is returned when pagination parameters are invalid
	ErrComparisonPhotoInvalidPagination = errors.New("invalid pagination parameters")

	// ErrComparisonPhotoInternalError is returned when an internal error occurs
	ErrComparisonPhotoInternalError = errors.New("internal comparison photo system error")

	// ErrComparisonPhotoInvalidFileType is returned when the uploaded file is not a valid image
	ErrComparisonPhotoInvalidFileType = errors.New("invalid file type for comparison photo")

	// ErrComparisonPhotoFileTooLarge is returned when the uploaded file exceeds size limits
	ErrComparisonPhotoFileTooLarge = errors.New("comparison photo file is too large")
)
