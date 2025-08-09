package models

// User represents a user in the database.
type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Role         string // "student" or "admin"
}
