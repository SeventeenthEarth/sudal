package entity

import (
	"time"

	"github.com/google/uuid"
)

// CandyTransactionType represents the type of candy transaction
type CandyTransactionType string

const (
	// Credit transactions (positive amounts)
	CandyTransactionTypePurchased            CandyTransactionType = "PURCHASED"
	CandyTransactionTypeEarnedQuizCompletion CandyTransactionType = "EARNED_QUIZ_COMPLETION"
	CandyTransactionTypeEarnedDaily          CandyTransactionType = "EARNED_DAILY_BONUS"
	CandyTransactionTypeEarnedReferral       CandyTransactionType = "EARNED_REFERRAL"
	CandyTransactionTypeRefund               CandyTransactionType = "REFUND"

	// Debit transactions (negative amounts)
	CandyTransactionTypeUsedComparisonFee  CandyTransactionType = "USED_COMPARISON_FEE"
	CandyTransactionTypeContributedToPot   CandyTransactionType = "CONTRIBUTED_TO_POT"
	CandyTransactionTypeUsedPowerUp        CandyTransactionType = "USED_POWER_UP"
	CandyTransactionTypeUsedPremiumFeature CandyTransactionType = "USED_PREMIUM_FEATURE"
)

// IsCredit returns true if the transaction type represents a credit (earning/receiving candy)
func (t CandyTransactionType) IsCredit() bool {
	switch t {
	case CandyTransactionTypePurchased,
		CandyTransactionTypeEarnedQuizCompletion,
		CandyTransactionTypeEarnedDaily,
		CandyTransactionTypeEarnedReferral,
		CandyTransactionTypeRefund:
		return true
	default:
		return false
	}
}

// IsDebit returns true if the transaction type represents a debit (spending candy)
func (t CandyTransactionType) IsDebit() bool {
	return !t.IsCredit()
}

// String returns the string representation of the transaction type
func (t CandyTransactionType) String() string {
	return string(t)
}

// CandyTransactionTypeStats represents statistics for a specific transaction type
type CandyTransactionTypeStats struct {
	Count  int64 `json:"count"`
	Amount int64 `json:"amount"`
}

// CandyTransaction represents a virtual currency transaction in the Social Quiz Platform
// This domain model encapsulates all candy transaction-related data and business rules
type CandyTransaction struct {
	// ID is the unique identifier for the transaction (BIGSERIAL for performance)
	ID int64 `json:"id"`

	// UserID is the foreign key reference to the user involved in the transaction
	UserID uuid.UUID `json:"user_id"`

	// Type is the transaction type (PURCHASED, EARNED_QUIZ_COMPLETION, etc.)
	Type CandyTransactionType `json:"type"`

	// Amount is the transaction amount (positive for credits, negative for debits)
	Amount int `json:"amount"`

	// BalanceAfterTransaction is the user's candy balance after this transaction
	// This provides an audit trail and helps with balance verification
	BalanceAfterTransaction int `json:"balance_after_transaction"`

	// Description is an optional description of the transaction
	Description *string `json:"description,omitempty"`

	// ReferenceID is an optional reference ID for linking to other entities
	// Examples: comparison_id, room_id, external_payment_id
	ReferenceID *string `json:"reference_id,omitempty"`

	// CreatedAt is the timestamp when the transaction was created (UTC)
	CreatedAt time.Time `json:"created_at"`
}

// NewCandyTransaction creates a new CandyTransaction with the provided parameters
// This constructor ensures required fields are set and provides sensible defaults
func NewCandyTransaction(userID uuid.UUID, transactionType CandyTransactionType, amount int, balanceAfter int) *CandyTransaction {
	return &CandyTransaction{
		UserID:                  userID,
		Type:                    transactionType,
		Amount:                  amount,
		BalanceAfterTransaction: balanceAfter,
		Description:             nil,
		ReferenceID:             nil,
		CreatedAt:               time.Now().UTC(),
	}
}

// SetDescription sets the transaction description
func (ct *CandyTransaction) SetDescription(description string) {
	ct.Description = &description
}

// SetReferenceID sets the reference ID for linking to other entities
func (ct *CandyTransaction) SetReferenceID(referenceID string) {
	ct.ReferenceID = &referenceID
}

// GetDescription returns the description or empty string if not set
func (ct *CandyTransaction) GetDescription() string {
	if ct.Description == nil {
		return ""
	}
	return *ct.Description
}

// GetReferenceID returns the reference ID or empty string if not set
func (ct *CandyTransaction) GetReferenceID() string {
	if ct.ReferenceID == nil {
		return ""
	}
	return *ct.ReferenceID
}

// IsCredit returns true if this transaction is a credit (positive amount)
func (ct *CandyTransaction) IsCredit() bool {
	return ct.Amount > 0
}

// IsDebit returns true if this transaction is a debit (negative amount)
func (ct *CandyTransaction) IsDebit() bool {
	return ct.Amount < 0
}

// GetAbsoluteAmount returns the absolute value of the transaction amount
func (ct *CandyTransaction) GetAbsoluteAmount() int {
	if ct.Amount < 0 {
		return -ct.Amount
	}
	return ct.Amount
}

// GetBalanceBeforeTransaction calculates the balance before this transaction
func (ct *CandyTransaction) GetBalanceBeforeTransaction() int {
	return ct.BalanceAfterTransaction - ct.Amount
}

// GetTransactionAge returns the duration since the transaction was created
func (ct *CandyTransaction) GetTransactionAge() time.Duration {
	return time.Since(ct.CreatedAt)
}

// IsRecentTransaction checks if the transaction was created recently (within the last hour)
func (ct *CandyTransaction) IsRecentTransaction() bool {
	return ct.GetTransactionAge() < time.Hour
}

// Validate performs basic validation on the candy transaction
func (ct *CandyTransaction) Validate() error {
	if ct.UserID == uuid.Nil {
		return ErrCandyTransactionInvalidUserID
	}
	if len(string(ct.Type)) == 0 {
		return ErrCandyTransactionTypeRequired
	}
	if ct.Amount == 0 {
		return ErrCandyTransactionAmountZero
	}
	if ct.BalanceAfterTransaction < 0 {
		return ErrCandyTransactionNegativeBalance
	}
	if ct.CreatedAt.IsZero() {
		return ErrCandyTransactionInvalidCreatedTime
	}

	// Validate transaction type consistency with amount
	if ct.Type.IsCredit() && ct.Amount <= 0 {
		return ErrCandyTransactionCreditAmountMustBePositive
	}
	if ct.Type.IsDebit() && ct.Amount >= 0 {
		return ErrCandyTransactionDebitAmountMustBeNegative
	}

	// Validate balance calculation
	expectedBalance := ct.GetBalanceBeforeTransaction() + ct.Amount
	if expectedBalance != ct.BalanceAfterTransaction {
		return ErrCandyTransactionInvalidBalanceCalculation
	}

	return nil
}
