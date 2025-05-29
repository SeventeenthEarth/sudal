package entity

import "errors"

// Tag Errors
var (
	// ErrTagNotFound is returned when a tag cannot be found
	ErrTagNotFound = errors.New("tag not found")

	// ErrTagAlreadyExists is returned when attempting to create a tag that already exists
	ErrTagAlreadyExists = errors.New("tag already exists")

	// ErrTagNameRequired is returned when a tag name is empty
	ErrTagNameRequired = errors.New("tag name is required")

	// ErrTagNameTooLong is returned when a tag name exceeds the maximum length
	ErrTagNameTooLong = errors.New("tag name is too long (maximum 50 characters)")

	// ErrTagDescriptionTooLong is returned when a tag description exceeds the maximum length
	ErrTagDescriptionTooLong = errors.New("tag description is too long (maximum 255 characters)")

	// ErrTagInvalidColor is returned when a tag color is not a valid hex color
	ErrTagInvalidColor = errors.New("tag color must be a valid hex color code (e.g., #FF5733)")

	// ErrTagInvalidID is returned when a tag ID is invalid
	ErrTagInvalidID = errors.New("invalid tag ID")

	// ErrTagInvalidPagination is returned when pagination parameters are invalid
	ErrTagInvalidPagination = errors.New("invalid pagination parameters")

	// ErrTagPermissionDenied is returned when a user doesn't have permission to perform an action
	ErrTagPermissionDenied = errors.New("permission denied for tag operation")

	// ErrTagInternalError is returned when an internal error occurs
	ErrTagInternalError = errors.New("internal tag system error")

	// ErrTagInUse is returned when attempting to delete a tag that is still in use
	ErrTagInUse = errors.New("tag is still in use and cannot be deleted")
)
