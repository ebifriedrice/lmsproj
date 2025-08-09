package database

import (
	"database/sql"
	"lms/internal/models"
	"time"

	"github.com/google/uuid"
)

// --- Course Functions ---

// CreateCourse creates a new course in the database.
func CreateCourse(db *sql.DB, title, description string) (*models.Course, error) {
	result, err := db.Exec("INSERT INTO courses (title, description) VALUES (?, ?)", title, description)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &models.Course{ID: id, Title: title, Description: description}, nil
}

// GetCourse retrieves a single course by its ID.
func GetCourse(db *sql.DB, id int64) (*models.Course, error) {
	row := db.QueryRow("SELECT id, title, description FROM courses WHERE id = ?", id)
	course := &models.Course{}
	err := row.Scan(&course.ID, &course.Title, &course.Description)
	if err != nil {
		return nil, err // Could be sql.ErrNoRows
	}
	return course, nil
}

// GetAllCourses retrieves all courses from the database.
func GetAllCourses(db *sql.DB) ([]*models.Course, error) {
	rows, err := db.Query("SELECT id, title, description FROM courses")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []*models.Course
	for rows.Next() {
		course := &models.Course{}
		if err := rows.Scan(&course.ID, &course.Title, &course.Description); err != nil {
			return nil, err
		}
		courses = append(courses, course)
	}
	return courses, nil
}

// --- Lesson Functions ---

// CreateLesson creates a new lesson for a course.
func CreateLesson(db *sql.DB, courseID int64, title string, position int) (*models.Lesson, error) {
	result, err := db.Exec("INSERT INTO lessons (course_id, title, position) VALUES (?, ?, ?)", courseID, title, position)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &models.Lesson{ID: id, CourseID: courseID, Title: title, Position: position}, nil
}

// GetLessonsForCourse retrieves all lessons for a given course, ordered by position.
func GetLessonsForCourse(db *sql.DB, courseID int64) ([]*models.Lesson, error) {
	rows, err := db.Query("SELECT id, course_id, title, position FROM lessons WHERE course_id = ? ORDER BY position ASC", courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lessons []*models.Lesson
	for rows.Next() {
		lesson := &models.Lesson{}
		if err := rows.Scan(&lesson.ID, &lesson.CourseID, &lesson.Title, &lesson.Position); err != nil {
			return nil, err
		}
		lessons = append(lessons, lesson)
	}
	return lessons, nil
}

// --- Enrollment Functions ---

// EnrollStudentInCourse enrolls a student in a course.
func EnrollStudentInCourse(db *sql.DB, userID, courseID int64) error {
	_, err := db.Exec("INSERT INTO enrollments (user_id, course_id) VALUES (?, ?)", userID, courseID)
	return err
}

// GetEnrolledCoursesForStudent retrieves all courses a student is enrolled in.
func GetEnrolledCoursesForStudent(db *sql.DB, userID int64) ([]*models.Course, error) {
	rows, err := db.Query(`
		SELECT c.id, c.title, c.description
		FROM courses c
		JOIN enrollments e ON c.id = e.course_id
		WHERE e.user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []*models.Course
	for rows.Next() {
		course := &models.Course{}
		if err := rows.Scan(&course.ID, &course.Title, &course.Description); err != nil {
			return nil, err
		}
		courses = append(courses, course)
	}
	return courses, nil
}


// --- Certificate Functions ---

// CreateCertificate generates a new unique certificate for a user and course.
func CreateCertificate(db *sql.DB, userID, courseID int64) (*models.Certificate, error) {
	token := uuid.New().String()
	result, err := db.Exec("INSERT INTO certificates (user_id, course_id, token) VALUES (?, ?, ?)", userID, courseID, token)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// We need to get the issued_at timestamp from the database.
	// Let's retrieve the certificate we just created.
	cert := &models.Certificate{}
	row := db.QueryRow("SELECT id, user_id, course_id, token, issued_at FROM certificates WHERE id = ?", id)
	err = row.Scan(&cert.ID, &cert.UserID, &cert.CourseID, &cert.Token, &cert.IssuedAt)
	if err != nil {
		return nil, err
	}
	return cert, nil
}

// --- Lesson Completion Functions ---

// MarkLessonAsComplete records that a user has completed a lesson.
func MarkLessonAsComplete(db *sql.DB, userID, lessonID int64) error {
	_, err := db.Exec("INSERT INTO lesson_completions (user_id, lesson_id) VALUES (?, ?)", userID, lessonID)
	return err
}

// GetCompletedLessonsForUser returns a map of completed lesson IDs for a user in a specific course.
func GetCompletedLessonsForUser(db *sql.DB, userID, courseID int64) (map[int64]bool, error) {
	rows, err := db.Query(`
		SELECT lc.lesson_id
		FROM lesson_completions lc
		JOIN lessons l ON lc.lesson_id = l.id
		WHERE lc.user_id = ? AND l.course_id = ?`, userID, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	completed := make(map[int64]bool)
	for rows.Next() {
		var lessonID int64
		if err := rows.Scan(&lessonID); err != nil {
			return nil, err
		}
		completed[lessonID] = true
	}
	return completed, nil
}

// IsCourseComplete checks if a user has completed all lessons in a course.
func IsCourseComplete(db *sql.DB, userID, courseID int64) (bool, error) {
	// Get all lesson IDs for the course.
	rows, err := db.Query("SELECT id FROM lessons WHERE course_id = ?", courseID)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	allLessons := make(map[int64]bool)
	for rows.Next() {
		var lessonID int64
		if err := rows.Scan(&lessonID); err != nil {
			return false, err
		}
		allLessons[lessonID] = true
	}
	if len(allLessons) == 0 {
		return false, nil // A course with no lessons cannot be completed.
	}

	// Get all completed lesson IDs for the user in this course.
	completedLessons, err := GetCompletedLessonsForUser(db, userID, courseID)
	if err != nil {
		return false, err
	}

	// Check if the sets are equal.
	if len(allLessons) != len(completedLessons) {
		return false, nil
	}

	for lessonID := range allLessons {
		if !completedLessons[lessonID] {
			return false, nil
		}
	}

	return true, nil
}

func GetLesson(db *sql.DB, id int64) (*models.Lesson, error) {
	row := db.QueryRow("SELECT id, course_id, title, position FROM lessons WHERE id = ?", id)
	lesson := &models.Lesson{}
	err := row.Scan(&lesson.ID, &lesson.CourseID, &lesson.Title, &lesson.Position)
	if err != nil {
		return nil, err
	}
	return lesson, nil
}

// IsLessonComplete checks if a user has completed a specific lesson.
func IsLessonComplete(db *sql.DB, userID, lessonID int64) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM lesson_completions WHERE user_id = ? AND lesson_id = ?)", userID, lessonID).Scan(&exists)
	return exists, err
}

// CertificateDetails contains all info for displaying a certificate.
type CertificateDetails struct {
	Token       string
	IssuedAt    time.Time
	StudentName string
	CourseTitle string
}

func GetCertificateDetailsByToken(db *sql.DB, token string) (*CertificateDetails, error) {
	row := db.QueryRow(`
        SELECT c.token, c.issued_at, u.username, co.title
        FROM certificates c
        JOIN users u ON c.user_id = u.id
        JOIN courses co ON c.course_id = co.id
        WHERE c.token = ?`, token)

	details := &CertificateDetails{}
	err := row.Scan(&details.Token, &details.IssuedAt, &details.StudentName, &details.CourseTitle)
	if err != nil {
		return nil, err
	}
	return details, nil
}
