package entity

import "errors"

// Candy Transaction Errors
var (
	// ErrCandyTransactionNotFound is returned when a candy transaction cannot be found
	ErrCandyTransactionNotFound = errors.New("candy transaction not found")

	// ErrCandyTransactionAlreadyExists is returned when attempting to create a transaction that already exists
	ErrCandyTransactionAlreadyExists = errors.New("candy transaction already exists")

	// ErrCandyTransactionInvalidID is returned when a transaction ID is invalid
	ErrCandyTransactionInvalidID = errors.New("invalid candy transaction ID")

	// ErrCandyTransactionInvalidUserID is returned when the user ID is invalid
	ErrCandyTransactionInvalidUserID = errors.New("invalid user ID for candy transaction")

	// ErrCandyTransactionTypeRequired is returned when a transaction type is empty
	ErrCandyTransactionTypeRequired = errors.New("candy transaction type is required")

	// ErrCandyTransactionInvalidType is returned when a transaction type is invalid
	ErrCandyTransactionInvalidType = errors.New("invalid candy transaction type")

	// ErrCandyTransactionAmountZero is returned when a transaction amount is zero
	ErrCandyTransactionAmountZero = errors.New("candy transaction amount cannot be zero")

	// ErrCandyTransactionNegativeBalance is returned when a balance after transaction is negative
	ErrCandyTransactionNegativeBalance = errors.New("candy balance cannot be negative")

	// ErrCandyTransactionInvalidCreatedTime is returned when the created time is invalid
	ErrCandyTransactionInvalidCreatedTime = errors.New("invalid candy transaction created time")

	// ErrCandyTransactionCreditAmountMustBePositive is returned when a credit transaction has non-positive amount
	ErrCandyTransactionCreditAmountMustBePositive = errors.New("credit candy transaction amount must be positive")

	// ErrCandyTransactionDebitAmountMustBeNegative is returned when a debit transaction has non-negative amount
	ErrCandyTransactionDebitAmountMustBeNegative = errors.New("debit candy transaction amount must be negative")

	// ErrCandyTransactionInvalidBalanceCalculation is returned when balance calculation doesn't match
	ErrCandyTransactionInvalidBalanceCalculation = errors.New("candy transaction balance calculation is invalid")

	// ErrCandyTransactionInsufficientFunds is returned when a user doesn't have enough candy for a transaction
	ErrCandyTransactionInsufficientFunds = errors.New("insufficient candy balance for transaction")

	// ErrCandyTransactionDuplicateReference is returned when a reference ID is already used
	ErrCandyTransactionDuplicateReference = errors.New("candy transaction reference ID already exists")

	// ErrCandyTransactionInvalidPagination is returned when pagination parameters are invalid
	ErrCandyTransactionInvalidPagination = errors.New("invalid pagination parameters")

	// ErrCandyTransactionPermissionDenied is returned when a user doesn't have permission to perform an action
	ErrCandyTransactionPermissionDenied = errors.New("permission denied for candy transaction operation")

	// ErrCandyTransactionInternalError is returned when an internal error occurs
	ErrCandyTransactionInternalError = errors.New("internal candy transaction system error")

	// ErrCandyTransactionSystemMaintenance is returned when the candy system is under maintenance
	ErrCandyTransactionSystemMaintenance = errors.New("candy transaction system is under maintenance")

	// ErrCandyTransactionRateLimitExceeded is returned when rate limits are exceeded
	ErrCandyTransactionRateLimitExceeded = errors.New("candy transaction rate limit exceeded")
)
