package entity

import "errors"

// Quiz Errors
var (
	// ErrQuizNotFound is returned when a quiz cannot be found
	ErrQuizNotFound = errors.New("quiz not found")

	// ErrQuizAlreadyExists is returned when attempting to create a quiz that already exists
	ErrQuizAlreadyExists = errors.New("quiz already exists")

	// ErrQuizTextRequired is returned when a quiz text is empty
	ErrQuizTextRequired = errors.New("quiz text is required")

	// ErrQuizOptionARequired is returned when option A is empty
	ErrQuizOptionARequired = errors.New("quiz option A is required")

	// ErrQuizOptionBRequired is returned when option B is empty
	ErrQuizOptionBRequired = errors.New("quiz option B is required")

	// ErrQuizOptionATooLong is returned when option A exceeds the maximum length
	ErrQuizOptionATooLong = errors.New("quiz option A is too long (maximum 255 characters)")

	// ErrQuizOptionBTooLong is returned when option B exceeds the maximum length
	ErrQuizOptionBTooLong = errors.New("quiz option B is too long (maximum 255 characters)")

	// ErrQuizInvalidOrder is returned when the quiz order is invalid
	ErrQuizInvalidOrder = errors.New("quiz order must be positive")

	// ErrQuizInvalidQuizSetID is returned when the quiz set ID is invalid
	ErrQuizInvalidQuizSetID = errors.New("invalid quiz set ID for quiz")

	// ErrQuizInvalidID is returned when a quiz ID is invalid
	ErrQuizInvalidID = errors.New("invalid quiz ID")

	// ErrQuizOrderConflict is returned when a quiz order conflicts with existing quizzes
	ErrQuizOrderConflict = errors.New("quiz order conflicts with existing quiz")

	// ErrQuizInvalidPagination is returned when pagination parameters are invalid
	ErrQuizInvalidPagination = errors.New("invalid pagination parameters")

	// ErrQuizPermissionDenied is returned when a user doesn't have permission to perform an action
	ErrQuizPermissionDenied = errors.New("permission denied for quiz operation")

	// ErrQuizInternalError is returned when an internal error occurs
	ErrQuizInternalError = errors.New("internal quiz system error")
)
