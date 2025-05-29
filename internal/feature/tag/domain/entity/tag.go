package entity

import "time"

// Tag represents a categorization tag for quiz sets in the Social Quiz Platform
// This domain model encapsulates all tag-related data and business rules
type Tag struct {
	// ID is the unique identifier for the tag (BIGSERIAL for performance)
	ID int64 `json:"id"`

	// Name is the tag name (required, unique)
	Name string `json:"name"`

	// Description is an optional description of what this tag represents
	Description *string `json:"description,omitempty"`

	// Color is an optional hex color code for UI display (e.g., "#FF5733")
	Color *string `json:"color,omitempty"`

	// CreatedAt is the timestamp when the tag was created (UTC)
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the timestamp when the tag was last updated (UTC)
	UpdatedAt time.Time `json:"updated_at"`
}

// NewTag creates a new Tag with the provided name and optional description
// This constructor ensures required fields are set and provides sensible defaults
func NewTag(name string, description *string) *Tag {
	now := time.Now().UTC()
	return &Tag{
		Name:        name,
		Description: description,
		Color:       nil, // No default color
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// UpdateName updates the tag's name and sets the updated timestamp
func (t *Tag) UpdateName(name string) {
	t.Name = name
	t.UpdatedAt = time.Now().UTC()
}

// UpdateDescription updates the tag's description and sets the updated timestamp
func (t *Tag) UpdateDescription(description *string) {
	t.Description = description
	t.UpdatedAt = time.Now().UTC()
}

// UpdateColor updates the tag's color and sets the updated timestamp
func (t *Tag) UpdateColor(color *string) {
	t.Color = color
	t.UpdatedAt = time.Now().UTC()
}

// GetDescription returns the description or empty string if not set
func (t *Tag) GetDescription() string {
	if t.Description == nil {
		return ""
	}
	return *t.Description
}

// GetColor returns the color or empty string if not set
func (t *Tag) GetColor() string {
	if t.Color == nil {
		return ""
	}
	return *t.Color
}

// HasDescription checks if the tag has a description
func (t *Tag) HasDescription() bool {
	return t.Description != nil && len(*t.Description) > 0
}

// HasColor checks if the tag has a color
func (t *Tag) HasColor() bool {
	return t.Color != nil && len(*t.Color) > 0
}

// SetDescription sets the tag's description
func (t *Tag) SetDescription(description string) {
	t.Description = &description
	t.UpdatedAt = time.Now().UTC()
}

// SetColor sets the tag's color
func (t *Tag) SetColor(color string) {
	t.Color = &color
	t.UpdatedAt = time.Now().UTC()
}

// ClearDescription removes the tag's description
func (t *Tag) ClearDescription() {
	t.Description = nil
	t.UpdatedAt = time.Now().UTC()
}

// ClearColor removes the tag's color
func (t *Tag) ClearColor() {
	t.Color = nil
	t.UpdatedAt = time.Now().UTC()
}

// Validate performs basic validation on the tag
func (t *Tag) Validate() error {
	if len(t.Name) == 0 {
		return ErrTagNameRequired
	}
	if len(t.Name) > 50 {
		return ErrTagNameTooLong
	}
	if t.Description != nil && len(*t.Description) > 255 {
		return ErrTagDescriptionTooLong
	}
	if t.Color != nil && !isValidHexColor(*t.Color) {
		return ErrTagInvalidColor
	}
	return nil
}

// isValidHexColor validates if a string is a valid hex color code
func isValidHexColor(color string) bool {
	if len(color) != 7 {
		return false
	}
	if color[0] != '#' {
		return false
	}
	for i := 1; i < 7; i++ {
		c := color[i]
		if (c < '0' || c > '9') && (c < 'A' || c > 'F') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}
