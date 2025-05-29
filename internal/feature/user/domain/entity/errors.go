package entity

import "errors"

// User domain errors
// These sentinel errors provide consistent error handling across the user domain

// ErrUserNotFound is returned when a user cannot be found by the specified criteria
var ErrUserNotFound = errors.New("user not found")

// ErrUserAlreadyExists is returned when attempting to create a user that already exists
var ErrUserAlreadyExists = errors.New("user already exists")

// ErrInvalidFirebaseUID is returned when the provided Firebase UID is invalid
var ErrInvalidFirebaseUID = errors.New("invalid firebase UID")

// ErrInvalidAuthProvider is returned when the provided auth provider is invalid
var ErrInvalidAuthProvider = errors.New("invalid auth provider")

// ErrInvalidDisplayName is returned when the provided display name is invalid
var ErrInvalidDisplayName = errors.New("invalid display name")

// ErrInvalidAmount is returned when an invalid amount is provided for candy operations
var ErrInvalidAmount = errors.New("invalid amount: must be non-negative")

// ErrInsufficientBalance is returned when a user doesn't have enough candy balance
var ErrInsufficientBalance = errors.New("insufficient candy balance")

// ErrInvalidUserID is returned when the provided user ID is invalid
var ErrInvalidUserID = errors.New("invalid user ID")
