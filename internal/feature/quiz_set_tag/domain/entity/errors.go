package entity

import "errors"

// Quiz Set Tag Association Errors
var (
	// ErrQuizSetTagNotFound is returned when a quiz set tag association cannot be found
	ErrQuizSetTagNotFound = errors.New("quiz set tag association not found")

	// ErrQuizSetTagAlreadyExists is returned when attempting to create a quiz set tag association that already exists
	ErrQuizSetTagAlreadyExists = errors.New("quiz set tag association already exists")

	// ErrQuizSetTagInvalidQuizSetID is returned when the quiz set ID is invalid
	ErrQuizSetTagInvalidQuizSetID = errors.New("invalid quiz set ID for tag association")

	// ErrQuizSetTagInvalidTagID is returned when the tag ID is invalid
	ErrQuizSetTagInvalidTagID = errors.New("invalid tag ID for quiz set association")

	// ErrQuizSetTagInvalidAssignedTime is returned when the assigned time is invalid
	ErrQuizSetTagInvalidAssignedTime = errors.New("invalid assigned time for quiz set tag association")

	// ErrQuizSetTagInvalidQuizSetTitle is returned when the quiz set title is invalid
	ErrQuizSetTagInvalidQuizSetTitle = errors.New("invalid quiz set title for tag association")

	// ErrQuizSetTagInvalidTagName is returned when the tag name is invalid
	ErrQuizSetTagInvalidTagName = errors.New("invalid tag name for quiz set association")

	// ErrQuizSetTagInvalidPagination is returned when pagination parameters are invalid
	ErrQuizSetTagInvalidPagination = errors.New("invalid pagination parameters")

	// ErrQuizSetTagPermissionDenied is returned when a user doesn't have permission to perform an action
	ErrQuizSetTagPermissionDenied = errors.New("permission denied for quiz set tag operation")

	// ErrQuizSetTagInternalError is returned when an internal error occurs
	ErrQuizSetTagInternalError = errors.New("internal quiz set tag system error")

	// ErrQuizSetTagBulkOperationFailed is returned when a bulk operation fails
	ErrQuizSetTagBulkOperationFailed = errors.New("quiz set tag bulk operation failed")

	// ErrQuizSetTagInvalidBulkData is returned when bulk operation data is invalid
	ErrQuizSetTagInvalidBulkData = errors.New("invalid data for quiz set tag bulk operation")
)
