package entity

import (
	"time"

	"github.com/google/uuid"
)

// ComparisonParticipant represents a user's participation in a comparison event
// This domain model encapsulates all participant-related data and business rules
type ComparisonParticipant struct {
	// ID is the unique identifier for the participant record (BIGSERIAL for performance)
	ID int64 `json:"id"`

	// ComparisonID is the foreign key reference to the comparison event
	ComparisonID int64 `json:"comparison_id"`

	// UserID is the foreign key reference to the participating user
	UserID uuid.UUID `json:"user_id"`

	// QuizResultID is the foreign key reference to the quiz result being compared
	// This represents the specific quiz attempt the user is bringing to the comparison
	QuizResultID int64 `json:"quiz_result_id"`

	// PersonalMemo is an optional personal note from the user about this comparison
	PersonalMemo *string `json:"personal_memo,omitempty"`

	// CreatedAt is the timestamp when the user joined the comparison (UTC)
	CreatedAt time.Time `json:"created_at"`
}

// NewComparisonParticipant creates a new ComparisonParticipant with the provided parameters
// This constructor ensures required fields are set and provides sensible defaults
func NewComparisonParticipant(comparisonID int64, userID uuid.UUID, quizResultID int64) *ComparisonParticipant {
	return &ComparisonParticipant{
		ComparisonID: comparisonID,
		UserID:       userID,
		QuizResultID: quizResultID,
		PersonalMemo: nil, // No memo by default
		CreatedAt:    time.Now().UTC(),
	}
}

// UpdatePersonalMemo updates the participant's personal memo
func (cp *ComparisonParticipant) UpdatePersonalMemo(memo *string) {
	cp.PersonalMemo = memo
}

// SetPersonalMemo sets the participant's personal memo with a string value
func (cp *ComparisonParticipant) SetPersonalMemo(memo string) {
	cp.PersonalMemo = &memo
}

// ClearPersonalMemo removes the participant's personal memo
func (cp *ComparisonParticipant) ClearPersonalMemo() {
	cp.PersonalMemo = nil
}

// HasPersonalMemo checks if the participant has a personal memo
func (cp *ComparisonParticipant) HasPersonalMemo() bool {
	return cp.PersonalMemo != nil && len(*cp.PersonalMemo) > 0
}

// GetPersonalMemo returns the personal memo or empty string if not set
func (cp *ComparisonParticipant) GetPersonalMemo() string {
	if cp.PersonalMemo == nil {
		return ""
	}
	return *cp.PersonalMemo
}

// GetParticipationAge returns the duration since the user joined the comparison
func (cp *ComparisonParticipant) GetParticipationAge() time.Duration {
	return time.Since(cp.CreatedAt)
}

// IsRecentParticipant checks if the user joined recently (within the last hour)
func (cp *ComparisonParticipant) IsRecentParticipant() bool {
	return cp.GetParticipationAge() < time.Hour
}

// Validate performs basic validation on the comparison participant
func (cp *ComparisonParticipant) Validate() error {
	if cp.ComparisonID <= 0 {
		return ErrComparisonParticipantInvalidComparisonID
	}
	if cp.UserID == uuid.Nil {
		return ErrComparisonParticipantInvalidUserID
	}
	if cp.QuizResultID <= 0 {
		return ErrComparisonParticipantInvalidQuizResultID
	}
	if cp.CreatedAt.IsZero() {
		return ErrComparisonParticipantInvalidCreatedTime
	}
	return nil
}
