package database

import (
	"database/sql"
	"encoding/json"
	"lms/internal/models"
)

// --- Video Functions ---

func CreateVideo(db *sql.DB, lessonID int64, title, url string) (*models.Video, error) {
	result, err := db.Exec("INSERT INTO videos (lesson_id, title, video_url) VALUES (?, ?, ?)", lessonID, title, url)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &models.Video{ID: id, LessonID: lessonID, Title: title, VideoURL: url}, nil
}

func GetVideoByLessonID(db *sql.DB, lessonID int64) (*models.Video, error) {
	row := db.QueryRow("SELECT id, lesson_id, title, video_url FROM videos WHERE lesson_id = ?", lessonID)
	video := &models.Video{}
	err := row.Scan(&video.ID, &video.LessonID, &video.Title, &video.VideoURL)
	if err != nil {
		return nil, err
	}
	return video, nil
}

// --- Text Functions ---

func CreateText(db *sql.DB, lessonID int64, title, content string) (*models.Text, error) {
	result, err := db.Exec("INSERT INTO texts (lesson_id, title, content) VALUES (?, ?, ?)", lessonID, title, content)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &models.Text{ID: id, LessonID: lessonID, Title: title, Content: content}, nil
}

func GetTextByLessonID(db *sql.DB, lessonID int64) (*models.Text, error) {
	row := db.QueryRow("SELECT id, lesson_id, title, content FROM texts WHERE lesson_id = ?", lessonID)
	text := &models.Text{}
	err := row.Scan(&text.ID, &text.LessonID, &text.Title, &text.Content)
	if err != nil {
		return nil, err
	}
	return text, nil
}

// --- MCQ Functions ---

func CreateMCQ(db *sql.DB, lessonID int64, question string, options []string, correctOptionIndex int) (*models.MCQ, error) {
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}

	result, err := db.Exec(
		"INSERT INTO mcqs (lesson_id, question, options, correct_option_index) VALUES (?, ?, ?, ?)",
		lessonID, question, string(optionsJSON), correctOptionIndex,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &models.MCQ{
		ID:                 id,
		LessonID:           lessonID,
		Question:           question,
		Options:            options,
		CorrectOptionIndex: correctOptionIndex,
	}, nil
}

func GetMCQByLessonID(db *sql.DB, lessonID int64) (*models.MCQ, error) {
	row := db.QueryRow("SELECT id, lesson_id, question, options, correct_option_index FROM mcqs WHERE lesson_id = ?", lessonID)
	mcq := &models.MCQ{}
	var optionsJSON string
	err := row.Scan(&mcq.ID, &mcq.LessonID, &mcq.Question, &optionsJSON, &mcq.CorrectOptionIndex)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(optionsJSON), &mcq.Options)
	if err != nil {
		return nil, err
	}
	return mcq, nil
}

func GetMCQByID(db *sql.DB, id int64) (*models.MCQ, error) {
	row := db.QueryRow("SELECT id, lesson_id, question, options, correct_option_index FROM mcqs WHERE id = ?", id)
	mcq := &models.MCQ{}
	var optionsJSON string
	err := row.Scan(&mcq.ID, &mcq.LessonID, &mcq.Question, &optionsJSON, &mcq.CorrectOptionIndex)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(optionsJSON), &mcq.Options)
	if err != nil {
		return nil, err
	}
	return mcq, nil
}

// --- MCQ Submission Functions ---

func SubmitMCQ(db *sql.DB, userID, mcqID int64, selectedOptionIndex int) (*models.MCQSubmission, error) {
	// First, get the correct answer to check if the submission is correct.
	row := db.QueryRow("SELECT correct_option_index FROM mcqs WHERE id = ?", mcqID)
	var correctOptionIndex int
	err := row.Scan(&correctOptionIndex)
	if err != nil {
		return nil, err
	}

	isCorrect := selectedOptionIndex == correctOptionIndex

	result, err := db.Exec(
		"INSERT INTO mcq_submissions (user_id, mcq_id, selected_option_index, is_correct) VALUES (?, ?, ?, ?)",
		userID, mcqID, selectedOptionIndex, isCorrect,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Retrieve the full submission record to get the timestamp.
	sub := &models.MCQSubmission{}
	row = db.QueryRow("SELECT id, user_id, mcq_id, selected_option_index, is_correct, submitted_at FROM mcq_submissions WHERE id = ?", id)
	err = row.Scan(&sub.ID, &sub.UserID, &sub.MCQID, &sub.SelectedOptionIndex, &sub.IsCorrect, &sub.SubmittedAt)
	if err != nil {
		return nil, err
	}

	return sub, nil
}
