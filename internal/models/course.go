package models

import "time"

// Course represents a course in the database.
type Course struct {
	ID          int64
	Title       string
	Description string
}

// Lesson represents a lesson within a course.
type Lesson struct {
	ID        int64
	CourseID  int64
	Title     string
	Position  int
}

// Certificate represents a certificate of completion for a course.
type Certificate struct {
	ID        int64
	UserID    int64
	CourseID  int64
	Token     string
	IssuedAt  time.Time
}
