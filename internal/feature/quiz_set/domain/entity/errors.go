package entity

import "errors"

// Quiz Set Errors
var (
	// ErrQuizSetNotFound is returned when a quiz set cannot be found
	ErrQuizSetNotFound = errors.New("quiz set not found")

	// ErrQuizSetAlreadyExists is returned when attempting to create a quiz set that already exists
	ErrQuizSetAlreadyExists = errors.New("quiz set already exists")

	// ErrQuizSetTitleRequired is returned when a quiz set title is empty
	ErrQuizSetTitleRequired = errors.New("quiz set title is required")

	// ErrQuizSetTitleTooLong is returned when a quiz set title exceeds the maximum length
	ErrQuizSetTitleTooLong = errors.New("quiz set title is too long (maximum 255 characters)")

	// ErrQuizSetInvalidID is returned when a quiz set ID is invalid
	ErrQuizSetInvalidID = errors.New("invalid quiz set ID")

	// ErrQuizSetInvalidPagination is returned when pagination parameters are invalid
	ErrQuizSetInvalidPagination = errors.New("invalid pagination parameters")

	// ErrQuizSetPermissionDenied is returned when a user doesn't have permission to perform an action
	ErrQuizSetPermissionDenied = errors.New("permission denied for quiz set operation")

	// ErrQuizSetInternalError is returned when an internal error occurs
	ErrQuizSetInternalError = errors.New("internal quiz set system error")
)
