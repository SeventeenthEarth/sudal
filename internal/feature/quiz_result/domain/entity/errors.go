package entity

import "errors"

// Quiz Result Errors
var (
	// ErrQuizResultNotFound is returned when a quiz result cannot be found
	ErrQuizResultNotFound = errors.New("quiz result not found")

	// ErrQuizResultAlreadyExists is returned when attempting to create a quiz result that already exists
	ErrQuizResultAlreadyExists = errors.New("quiz result already exists")

	// ErrQuizResultInvalidUserID is returned when the user ID is invalid
	ErrQuizResultInvalidUserID = errors.New("invalid user ID for quiz result")

	// ErrQuizResultInvalidQuizSetID is returned when the quiz set ID is invalid
	ErrQuizResultInvalidQuizSetID = errors.New("invalid quiz set ID for quiz result")

	// ErrQuizResultNoAnswers is returned when a quiz result has no answers
	ErrQuizResultNoAnswers = errors.New("quiz result must have at least one answer")

	// ErrQuizResultInvalidSubmissionTime is returned when the submission time is invalid
	ErrQuizResultInvalidSubmissionTime = errors.New("invalid quiz result submission time")

	// ErrQuizResultInvalidQuestionIndex is returned when accessing an invalid question index
	ErrQuizResultInvalidQuestionIndex = errors.New("invalid question index")

	// ErrQuizResultAnswerCountMismatch is returned when answer counts don't match
	ErrQuizResultAnswerCountMismatch = errors.New("answer count mismatch")

	// ErrQuizResultNoQuestions is returned when there are no questions to calculate score
	ErrQuizResultNoQuestions = errors.New("no questions available for score calculation")

	// ErrQuizResultInvalidID is returned when a quiz result ID is invalid
	ErrQuizResultInvalidID = errors.New("invalid quiz result ID")

	// ErrQuizResultInvalidPagination is returned when pagination parameters are invalid
	ErrQuizResultInvalidPagination = errors.New("invalid pagination parameters")

	// ErrQuizResultPermissionDenied is returned when a user doesn't have permission to perform an action
	ErrQuizResultPermissionDenied = errors.New("permission denied for quiz result operation")

	// ErrQuizResultInternalError is returned when an internal error occurs
	ErrQuizResultInternalError = errors.New("internal quiz result system error")
)
