package entity

import (
	"time"

	"github.com/google/uuid"
)

// QuizResult represents a user's attempt at a quiz in the Social Quiz Platform
// This domain model encapsulates all quiz result-related data and business rules
type QuizResult struct {
	// ID is the unique identifier for the quiz result (BIGSERIAL for performance)
	ID int64 `json:"id"`

	// UserID is the foreign key reference to the user who took the quiz
	UserID uuid.UUID `json:"user_id"`

	// QuizSetID is the foreign key reference to the quiz set that was taken
	QuizSetID int64 `json:"quiz_set_id"`

	// Answers is a slice of boolean values representing the user's choices
	// true = option A, false = option B
	// The order corresponds to the question order
	Answers []bool `json:"answers"`

	// SubmittedAt is the timestamp when the quiz was submitted (UTC)
	SubmittedAt time.Time `json:"submitted_at"`
}

// NewQuizResult creates a new QuizResult with the provided parameters
// This constructor ensures required fields are set and provides sensible defaults
func NewQuizResult(userID uuid.UUID, quizSetID int64, answers []bool) *QuizResult {
	return &QuizResult{
		UserID:      userID,
		QuizSetID:   quizSetID,
		Answers:     make([]bool, len(answers)),
		SubmittedAt: time.Now().UTC(),
	}
}

// SetAnswers sets the user's answers and copies the slice to prevent external modification
func (qr *QuizResult) SetAnswers(answers []bool) {
	qr.Answers = make([]bool, len(answers))
	copy(qr.Answers, answers)
}

// GetAnswer returns the answer for a specific question index (0-based)
func (qr *QuizResult) GetAnswer(questionIndex int) (bool, error) {
	if questionIndex < 0 || questionIndex >= len(qr.Answers) {
		return false, ErrQuizResultInvalidQuestionIndex
	}
	return qr.Answers[questionIndex], nil
}

// SetAnswer sets the answer for a specific question index (0-based)
func (qr *QuizResult) SetAnswer(questionIndex int, answer bool) error {
	if questionIndex < 0 || questionIndex >= len(qr.Answers) {
		return ErrQuizResultInvalidQuestionIndex
	}
	qr.Answers[questionIndex] = answer
	return nil
}

// GetAnswerCount returns the total number of answers
func (qr *QuizResult) GetAnswerCount() int {
	return len(qr.Answers)
}

// CalculateScore calculates the score by comparing answers with correct answers
// This method requires the correct answers to be provided
func (qr *QuizResult) CalculateScore(correctAnswers []bool) (int, error) {
	if len(qr.Answers) != len(correctAnswers) {
		return 0, ErrQuizResultAnswerCountMismatch
	}

	score := 0
	for i, userAnswer := range qr.Answers {
		if userAnswer == correctAnswers[i] {
			score++
		}
	}
	return score, nil
}

// GetScorePercentage calculates the score as a percentage
func (qr *QuizResult) GetScorePercentage(correctAnswers []bool) (float64, error) {
	score, err := qr.CalculateScore(correctAnswers)
	if err != nil {
		return 0, err
	}

	if len(correctAnswers) == 0 {
		return 0, ErrQuizResultNoQuestions
	}

	return float64(score) / float64(len(correctAnswers)) * 100, nil
}

// CompareAnswers compares this quiz result with another quiz result
// Returns a slice of booleans indicating which answers match
func (qr *QuizResult) CompareAnswers(other *QuizResult) ([]bool, error) {
	if len(qr.Answers) != len(other.Answers) {
		return nil, ErrQuizResultAnswerCountMismatch
	}

	matches := make([]bool, len(qr.Answers))
	for i := range qr.Answers {
		matches[i] = qr.Answers[i] == other.Answers[i]
	}
	return matches, nil
}

// GetSimilarityPercentage calculates the similarity percentage with another quiz result
func (qr *QuizResult) GetSimilarityPercentage(other *QuizResult) (float64, error) {
	matches, err := qr.CompareAnswers(other)
	if err != nil {
		return 0, err
	}

	if len(matches) == 0 {
		return 0, ErrQuizResultNoQuestions
	}

	matchCount := 0
	for _, match := range matches {
		if match {
			matchCount++
		}
	}

	return float64(matchCount) / float64(len(matches)) * 100, nil
}

// GetSubmissionAge returns the duration since the quiz was submitted
func (qr *QuizResult) GetSubmissionAge() time.Duration {
	return time.Since(qr.SubmittedAt)
}

// IsRecentSubmission checks if the quiz was submitted recently (within the last hour)
func (qr *QuizResult) IsRecentSubmission() bool {
	return qr.GetSubmissionAge() < time.Hour
}

// Validate performs basic validation on the quiz result
func (qr *QuizResult) Validate() error {
	if qr.UserID == uuid.Nil {
		return ErrQuizResultInvalidUserID
	}
	if qr.QuizSetID <= 0 {
		return ErrQuizResultInvalidQuizSetID
	}
	if len(qr.Answers) == 0 {
		return ErrQuizResultNoAnswers
	}
	if qr.SubmittedAt.IsZero() {
		return ErrQuizResultInvalidSubmissionTime
	}
	return nil
}

// IsComplete checks if the quiz result has answers for all questions
func (qr *QuizResult) IsComplete(expectedQuestionCount int) bool {
	return len(qr.Answers) == expectedQuestionCount
}
