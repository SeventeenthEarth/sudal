package entity

import "time"

// Comparison represents a comparison event where users compare their quiz results
// This domain model encapsulates all comparison-related data and business rules
type Comparison struct {
	// ID is the unique identifier for the comparison (BIGSERIAL for performance)
	ID int64 `json:"id"`

	// QuizSetID is the foreign key reference to the quiz set being compared
	QuizSetID int64 `json:"quiz_set_id"`

	// RoomID is the room identifier for grouping users in the comparison session
	RoomID string `json:"room_id"`

	// CreatedAt is the timestamp when the comparison was created (UTC)
	CreatedAt time.Time `json:"created_at"`
}

// NewComparison creates a new Comparison with the provided parameters
// This constructor ensures required fields are set and provides sensible defaults
func NewComparison(quizSetID int64, roomID string) *Comparison {
	return &Comparison{
		QuizSetID: quizSetID,
		RoomID:    roomID,
		CreatedAt: time.Now().UTC(),
	}
}

// GetAge returns the duration since the comparison was created
func (c *Comparison) GetAge() time.Duration {
	return time.Since(c.CreatedAt)
}

// IsRecentComparison checks if the comparison was created recently (within the last hour)
func (c *Comparison) IsRecentComparison() bool {
	return c.GetAge() < time.Hour
}

// IsActiveComparison checks if the comparison has recent activity (within the last 24 hours)
func (c *Comparison) IsActiveComparison() bool {
	return c.GetAge() < 24*time.Hour
}

// Validate performs basic validation on the comparison
func (c *Comparison) Validate() error {
	if c.QuizSetID <= 0 {
		return ErrComparisonInvalidQuizSetID
	}
	if len(c.RoomID) == 0 {
		return ErrComparisonRoomIDRequired
	}
	if len(c.RoomID) > 255 {
		return ErrComparisonRoomIDTooLong
	}
	if c.CreatedAt.IsZero() {
		return ErrComparisonInvalidCreatedTime
	}
	return nil
}
