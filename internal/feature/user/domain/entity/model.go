package entity

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the Social Quiz Platform
// This domain model encapsulates all user-related data and business rules
type User struct {
	// ID is the unique identifier for the user (UUID for security and distribution)
	ID uuid.UUID `json:"id"`

	// FirebaseUID is the Firebase authentication identifier (required, unique)
	// This links the user to their Firebase authentication account
	FirebaseUID string `json:"firebase_uid"`

	// DisplayName is the user's chosen display name (optional)
	// This is what other users see in the platform
	DisplayName *string `json:"display_name,omitempty"`

	// AvatarURL is the URL to the user's avatar image (optional)
	// Can be a Firebase Storage URL or external image URL
	AvatarURL *string `json:"avatar_url,omitempty"`

	// CandyBalance is the user's virtual currency balance
	// Used for purchasing quiz sets, power-ups, or other platform features
	CandyBalance int `json:"candy_balance"`

	// AuthProvider indicates which OAuth provider was used for authentication
	// Examples: "google", "apple", "facebook", "email"
	AuthProvider string `json:"auth_provider"`

	// CreatedAt is the timestamp when the user account was created (UTC)
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the timestamp when the user account was last updated (UTC)
	UpdatedAt time.Time `json:"updated_at"`
}

// NewUser creates a new User with the provided Firebase UID and auth provider
// This constructor ensures required fields are set and provides sensible defaults
func NewUser(firebaseUID, authProvider string) *User {
	now := time.Now().UTC()
	return &User{
		ID:           uuid.New(),
		FirebaseUID:  firebaseUID,
		DisplayName:  nil, // Will be set later by user
		AvatarURL:    nil, // Will be set later by user
		CandyBalance: 0,   // Start with zero balance
		AuthProvider: authProvider,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// UpdateDisplayName updates the user's display name and sets the updated timestamp
func (u *User) UpdateDisplayName(displayName string) {
	u.DisplayName = &displayName
	u.UpdatedAt = time.Now().UTC()
}

// UpdateAvatarURL updates the user's avatar URL and sets the updated timestamp
func (u *User) UpdateAvatarURL(avatarURL string) {
	u.AvatarURL = &avatarURL
	u.UpdatedAt = time.Now().UTC()
}

// AddCandy adds the specified amount to the user's candy balance
// Returns an error if the amount would result in a negative balance
func (u *User) AddCandy(amount int) error {
	if u.CandyBalance+amount < 0 {
		return ErrInsufficientBalance
	}
	u.CandyBalance += amount
	u.UpdatedAt = time.Now().UTC()
	return nil
}

// DeductCandy deducts the specified amount from the user's candy balance
// Returns an error if the user doesn't have sufficient balance
func (u *User) DeductCandy(amount int) error {
	if amount < 0 {
		return ErrInvalidAmount
	}
	if u.CandyBalance < amount {
		return ErrInsufficientBalance
	}
	u.CandyBalance -= amount
	u.UpdatedAt = time.Now().UTC()
	return nil
}

// HasSufficientBalance checks if the user has enough candy for a transaction
func (u *User) HasSufficientBalance(amount int) bool {
	return u.CandyBalance >= amount
}

// IsValidDisplayName checks if a display name meets the platform requirements
func IsValidDisplayName(displayName string) bool {
	if displayName == "" {
		return false
	}
	// Check length constraints (matching database constraint)
	if len(displayName) > 100 {
		return false
	}
	// Check for empty or whitespace-only names
	trimmed := strings.TrimSpace(displayName)
	return len(trimmed) > 0
}
