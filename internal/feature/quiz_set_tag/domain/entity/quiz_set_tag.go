package entity

import "time"

// QuizSetTag represents the many-to-many relationship between quiz sets and tags
// This domain model encapsulates the association between quiz sets and their tags
type QuizSetTag struct {
	// QuizSetID is the foreign key reference to the quiz set
	QuizSetID int64 `json:"quiz_set_id"`

	// TagID is the foreign key reference to the tag
	TagID int64 `json:"tag_id"`

	// AssignedAt is the timestamp when the tag was assigned to the quiz set (UTC)
	AssignedAt time.Time `json:"assigned_at"`
}

// NewQuizSetTag creates a new QuizSetTag association
func NewQuizSetTag(quizSetID, tagID int64) *QuizSetTag {
	return &QuizSetTag{
		QuizSetID:  quizSetID,
		TagID:      tagID,
		AssignedAt: time.Now().UTC(),
	}
}

// GetAssignmentAge returns the duration since the tag was assigned
func (qst *QuizSetTag) GetAssignmentAge() time.Duration {
	return time.Since(qst.AssignedAt)
}

// IsRecentAssignment checks if the tag was assigned recently (within the last hour)
func (qst *QuizSetTag) IsRecentAssignment() bool {
	return qst.GetAssignmentAge() < time.Hour
}

// Validate performs basic validation on the quiz set tag association
func (qst *QuizSetTag) Validate() error {
	if qst.QuizSetID <= 0 {
		return ErrQuizSetTagInvalidQuizSetID
	}
	if qst.TagID <= 0 {
		return ErrQuizSetTagInvalidTagID
	}
	if qst.AssignedAt.IsZero() {
		return ErrQuizSetTagInvalidAssignedTime
	}
	return nil
}

// QuizSetTagAssociation represents a simple association between a quiz set and tag for bulk operations
type QuizSetTagAssociation struct {
	QuizSetID int64 `json:"quiz_set_id"`
	TagID     int64 `json:"tag_id"`
}

// QuizSetTagDetail represents detailed information about a quiz set tag association
// This is a DTO-like struct used for displaying association information with related data
type QuizSetTagDetail struct {
	// QuizSetID is the unique identifier for the quiz set
	QuizSetID int64 `json:"quiz_set_id"`

	// QuizSetTitle is the title of the quiz set
	QuizSetTitle string `json:"quiz_set_title"`

	// TagID is the unique identifier for the tag
	TagID int64 `json:"tag_id"`

	// TagName is the name of the tag
	TagName string `json:"tag_name"`

	// TagDescription is the description of the tag
	TagDescription *string `json:"tag_description,omitempty"`

	// TagColor is the color of the tag
	TagColor *string `json:"tag_color,omitempty"`

	// AssignedAt is the timestamp when the tag was assigned
	AssignedAt time.Time `json:"assigned_at"`
}

// NewQuizSetTagDetail creates a new QuizSetTagDetail with the provided parameters
func NewQuizSetTagDetail(
	quizSetID int64,
	quizSetTitle string,
	tagID int64,
	tagName string,
	tagDescription, tagColor *string,
	assignedAt time.Time,
) *QuizSetTagDetail {
	return &QuizSetTagDetail{
		QuizSetID:      quizSetID,
		QuizSetTitle:   quizSetTitle,
		TagID:          tagID,
		TagName:        tagName,
		TagDescription: tagDescription,
		TagColor:       tagColor,
		AssignedAt:     assignedAt,
	}
}

// GetTagDescription returns the tag description or empty string if not set
func (qstd *QuizSetTagDetail) GetTagDescription() string {
	if qstd.TagDescription == nil {
		return ""
	}
	return *qstd.TagDescription
}

// GetTagColor returns the tag color or empty string if not set
func (qstd *QuizSetTagDetail) GetTagColor() string {
	if qstd.TagColor == nil {
		return ""
	}
	return *qstd.TagColor
}

// HasTagDescription checks if the tag has a description
func (qstd *QuizSetTagDetail) HasTagDescription() bool {
	return qstd.TagDescription != nil && len(*qstd.TagDescription) > 0
}

// HasTagColor checks if the tag has a color
func (qstd *QuizSetTagDetail) HasTagColor() bool {
	return qstd.TagColor != nil && len(*qstd.TagColor) > 0
}

// GetAssignmentAge returns the duration since the tag was assigned
func (qstd *QuizSetTagDetail) GetAssignmentAge() time.Duration {
	return time.Since(qstd.AssignedAt)
}

// IsRecentAssignment checks if the tag was assigned recently (within the last hour)
func (qstd *QuizSetTagDetail) IsRecentAssignment() bool {
	return qstd.GetAssignmentAge() < time.Hour
}

// Validate performs basic validation on the quiz set tag detail
func (qstd *QuizSetTagDetail) Validate() error {
	if qstd.QuizSetID <= 0 {
		return ErrQuizSetTagInvalidQuizSetID
	}
	if len(qstd.QuizSetTitle) == 0 {
		return ErrQuizSetTagInvalidQuizSetTitle
	}
	if qstd.TagID <= 0 {
		return ErrQuizSetTagInvalidTagID
	}
	if len(qstd.TagName) == 0 {
		return ErrQuizSetTagInvalidTagName
	}
	if qstd.AssignedAt.IsZero() {
		return ErrQuizSetTagInvalidAssignedTime
	}
	return nil
}
