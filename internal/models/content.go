package models

import "time"

// Video represents a video lecture content.
type Video struct {
	ID       int64
	LessonID int64
	Title    string
	VideoURL string
}

// Text represents a text content.
type Text struct {
	ID       int64
	LessonID int64
	Title    string
	Content  string
}

// MCQ represents a multiple choice question.
type MCQ struct {
	ID                 int64
	LessonID           int64
	Question           string
	Options            []string // Decoded from JSON
	CorrectOptionIndex int
}

// MCQSubmission represents a student's submission for an MCQ.
type MCQSubmission struct {
	ID                  int64
	UserID              int64
	MCQID               int64
	SelectedOptionIndex int
	IsCorrect           bool
	SubmittedAt         time.Time
}
