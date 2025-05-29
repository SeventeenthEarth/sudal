package entity

import "time"

// Quiz represents an individual quiz question in the Social Quiz Platform
// This domain model encapsulates all quiz-related data and business rules
type Quiz struct {
	// ID is the unique identifier for the quiz (BIGSERIAL for performance)
	ID int64 `json:"id"`

	// QuizSetID is the foreign key reference to the quiz set this quiz belongs to
	QuizSetID int64 `json:"quiz_set_id"`

	// Text is the quiz text displayed to users (required)
	Text string `json:"text"`

	// OptionA is the first answer option (required)
	OptionA string `json:"option_a"`

	// OptionB is the second answer option (required)
	OptionB string `json:"option_b"`

	// QuizOrder is the order of this quiz within the quiz set (required)
	// Used for displaying quizzes in a specific sequence
	QuizOrder int `json:"quiz_order"`

	// CreatedAt is the timestamp when the quiz was created (UTC)
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the timestamp when the quiz was last updated (UTC)
	UpdatedAt time.Time `json:"updated_at"`
}

// NewQuiz creates a new Quiz with the provided parameters
// This constructor ensures required fields are set and provides sensible defaults
func NewQuiz(quizSetID int64, text, optionA, optionB string, quizOrder int) *Quiz {
	now := time.Now().UTC()
	return &Quiz{
		QuizSetID: quizSetID,
		Text:      text,
		OptionA:   optionA,
		OptionB:   optionB,
		QuizOrder: quizOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UpdateText updates the quiz's text and sets the updated timestamp
func (q *Quiz) UpdateText(text string) {
	q.Text = text
	q.UpdatedAt = time.Now().UTC()
}

// UpdateOptionA updates the first option and sets the updated timestamp
func (q *Quiz) UpdateOptionA(optionA string) {
	q.OptionA = optionA
	q.UpdatedAt = time.Now().UTC()
}

// UpdateOptionB updates the second option and sets the updated timestamp
func (q *Quiz) UpdateOptionB(optionB string) {
	q.OptionB = optionB
	q.UpdatedAt = time.Now().UTC()
}

// UpdateOptions updates both options and sets the updated timestamp
func (q *Quiz) UpdateOptions(optionA, optionB string) {
	q.OptionA = optionA
	q.OptionB = optionB
	q.UpdatedAt = time.Now().UTC()
}

// UpdateQuizOrder updates the quiz's order and sets the updated timestamp
func (q *Quiz) UpdateQuizOrder(quizOrder int) {
	q.QuizOrder = quizOrder
	q.UpdatedAt = time.Now().UTC()
}

// Validate performs basic validation on the quiz
func (q *Quiz) Validate() error {
	if q.QuizSetID <= 0 {
		return ErrQuizInvalidQuizSetID
	}
	if len(q.Text) == 0 {
		return ErrQuizTextRequired
	}
	if len(q.OptionA) == 0 {
		return ErrQuizOptionARequired
	}
	if len(q.OptionB) == 0 {
		return ErrQuizOptionBRequired
	}
	if len(q.OptionA) > 255 {
		return ErrQuizOptionATooLong
	}
	if len(q.OptionB) > 255 {
		return ErrQuizOptionBTooLong
	}
	if q.QuizOrder < 1 {
		return ErrQuizInvalidOrder
	}
	return nil
}

// IsValidAnswer checks if the provided answer is valid (true for A, false for B)
func (q *Quiz) IsValidAnswer(answer bool) bool {
	// Both true (A) and false (B) are valid answers for binary choice questions
	return true
}

// GetAnswerText returns the text of the selected answer
func (q *Quiz) GetAnswerText(answer bool) string {
	if answer {
		return q.OptionA
	}
	return q.OptionB
}
