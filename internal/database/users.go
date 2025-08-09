package database

import (
	"database/sql"
	"errors"
	"lms/internal/models"

	"golang.org/x/crypto/bcrypt"
)

// ErrUserNotFound is returned when a user is not found in the database.
var ErrUserNotFound = errors.New("user not found")

// CreateUser hashes the password and inserts a new user into the database.
// It returns the newly created user.
func CreateUser(db *sql.DB, username, password, role string) (*models.User, error) {
	// Hash the password using bcrypt.
	// The second argument is the cost of hashing. DefaultCost is a good value.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Insert the new user into the database.
	result, err := db.Exec(
		"INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)",
		username,
		string(hashedPassword),
		role,
	)
	if err != nil {
		return nil, err
	}

	// Get the ID of the newly inserted user.
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Create a User model to return.
	user := &models.User{
		ID:           id,
		Username:     username,
		PasswordHash: string(hashedPassword),
		Role:         role,
	}

	return user, nil
}

// GetUserByUsername retrieves a user from the database by their username.
// It returns ErrUserNotFound if the user does not exist.
func GetUserByUsername(db *sql.DB, username string) (*models.User, error) {
	user := &models.User{}
	row := db.QueryRow("SELECT id, username, password_hash, role FROM users WHERE username = ?", username)

	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			// Use a custom error to indicate that the user was not found.
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// AuthenticateUser checks if a user's credentials are valid.
// It returns the user object on success.
func AuthenticateUser(db *sql.DB, username, password string) (*models.User, error) {
	// Retrieve the user from the database.
	user, err := GetUserByUsername(db, username)
	if err != nil {
		return nil, err // This will be ErrUserNotFound or a database error.
	}

	// Compare the provided password with the stored hash.
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		// If the passwords don't match, bcrypt returns an error.
		// We return the same error as user not found to prevent username enumeration.
		return nil, ErrUserNotFound
	}

	// Passwords match.
	return user, nil
}

// GetAllUsers retrieves all users from the database.
func GetAllUsers(db *sql.DB) ([]*models.User, error) {
	rows, err := db.Query("SELECT id, username, role FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		// Note: We are not selecting the password hash for security reasons.
		if err := rows.Scan(&user.ID, &user.Username, &user.Role); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// GetUserByID retrieves a single user by their ID.
func GetUserByID(db *sql.DB, id int64) (*models.User, error) {
	row := db.QueryRow("SELECT id, username, role FROM users WHERE id = ?", id)
	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Role)
	if err != nil {
		return nil, err // Could be sql.ErrNoRows
	}
	return user, nil
}
