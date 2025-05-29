package entity

import "errors"

// Comparison domain errors
var (
	// ErrComparisonNotFound is returned when a comparison is not found
	ErrComparisonNotFound = errors.New("comparison not found")

	// ErrComparisonAlreadyExists is returned when trying to create a comparison that already exists
	ErrComparisonAlreadyExists = errors.New("comparison already exists")

	// ErrComparisonInvalidID is returned when an invalid comparison ID is provided
	ErrComparisonInvalidID = errors.New("invalid comparison ID")

	// ErrComparisonInvalidQuizSetID is returned when an invalid quiz set ID is provided
	ErrComparisonInvalidQuizSetID = errors.New("invalid quiz set ID")

	// ErrComparisonRoomIDRequired is returned when room ID is empty
	ErrComparisonRoomIDRequired = errors.New("room ID is required")

	// ErrComparisonRoomIDTooLong is returned when room ID exceeds maximum length
	ErrComparisonRoomIDTooLong = errors.New("room ID is too long")

	// ErrComparisonInvalidCreatedTime is returned when created time is invalid
	ErrComparisonInvalidCreatedTime = errors.New("invalid created time")
)
