package entity

import "errors"

// Comparison Participant Errors
var (
	// ErrComparisonParticipantNotFound is returned when a comparison participant cannot be found
	ErrComparisonParticipantNotFound = errors.New("comparison participant not found")

	// ErrComparisonParticipantAlreadyExists is returned when attempting to create a participant that already exists
	ErrComparisonParticipantAlreadyExists = errors.New("comparison participant already exists")

	// ErrComparisonParticipantInvalidID is returned when a participant ID is invalid
	ErrComparisonParticipantInvalidID = errors.New("invalid comparison participant ID")

	// ErrComparisonParticipantInvalidComparisonID is returned when the comparison ID is invalid
	ErrComparisonParticipantInvalidComparisonID = errors.New("invalid comparison ID for participant")

	// ErrComparisonParticipantInvalidUserID is returned when the user ID is invalid
	ErrComparisonParticipantInvalidUserID = errors.New("invalid user ID for comparison participant")

	// ErrComparisonParticipantInvalidQuizResultID is returned when the quiz result ID is invalid
	ErrComparisonParticipantInvalidQuizResultID = errors.New("invalid quiz result ID for comparison participant")

	// ErrComparisonParticipantInvalidCreatedTime is returned when the created time is invalid
	ErrComparisonParticipantInvalidCreatedTime = errors.New("invalid comparison participant created time")

	// ErrComparisonParticipantDuplicateUser is returned when a user tries to participate twice in the same comparison
	ErrComparisonParticipantDuplicateUser = errors.New("user already participating in this comparison")

	// ErrComparisonParticipantInvalidPagination is returned when pagination parameters are invalid
	ErrComparisonParticipantInvalidPagination = errors.New("invalid pagination parameters")

	// ErrComparisonParticipantPermissionDenied is returned when a user doesn't have permission to perform an action
	ErrComparisonParticipantPermissionDenied = errors.New("permission denied for comparison participant operation")

	// ErrComparisonParticipantInternalError is returned when an internal error occurs
	ErrComparisonParticipantInternalError = errors.New("internal comparison participant system error")
)
