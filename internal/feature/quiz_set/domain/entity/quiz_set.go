package entity

import (
	"time"
)

// QuizSet represents a collection of quiz questions in the Social Quiz Platform
// This domain model encapsulates all quiz set-related data and business rules
type QuizSet struct {
	// ID is the unique identifier for the quiz set (BIGSERIAL for performance)
	ID int64 `json:"id"`

	// Title is the quiz set title displayed to users (required)
	Title string `json:"title"`

	// Description is an optional description of the quiz set
	Description *string `json:"description,omitempty"`

	// CreatedAt is the timestamp when the quiz set was created (UTC)
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the timestamp when the quiz set was last updated (UTC)
	UpdatedAt time.Time `json:"updated_at"`
}

// NewQuizSet creates a new QuizSet with the provided title and optional description
// This constructor ensures required fields are set and provides sensible defaults
func NewQuizSet(title string, description *string) *QuizSet {
	now := time.Now().UTC()
	return &QuizSet{
		Title:       title,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// UpdateTitle updates the quiz set's title and sets the updated timestamp
func (qs *QuizSet) UpdateTitle(title string) {
	qs.Title = title
	qs.UpdatedAt = time.Now().UTC()
}

// UpdateDescription updates the quiz set's description and sets the updated timestamp
func (qs *QuizSet) UpdateDescription(description *string) {
	qs.Description = description
	qs.UpdatedAt = time.Now().UTC()
}

// Validate performs basic validation on the quiz set
func (qs *QuizSet) Validate() error {
	if len(qs.Title) == 0 {
		return ErrQuizSetTitleRequired
	}
	if len(qs.Title) > 255 {
		return ErrQuizSetTitleTooLong
	}
	return nil
}
