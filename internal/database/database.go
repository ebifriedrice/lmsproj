package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"sort"

	_ "github.com/mattn/go-sqlite3" // Import the sqlite3 driver
)

// NewDB creates a new database connection.
func NewDB(dsn string) (*sql.DB, error) {
	// Open the database connection.
	// The DSN (Data Source Name) for SQLite is just the file path.
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	// Ping the database to verify the connection is alive.
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// ApplyMigrations reads all .up.sql files from a directory and applies them to the database.
func ApplyMigrations(db *sql.DB, dir string) error {
	// Find all migration files.
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return err
	}

	// Sort the files to ensure they are applied in the correct order.
	sort.Strings(files)

	// Loop through the files and execute them.
	for _, file := range files {
		// Read the content of the migration file.
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		// Execute the SQL script.
		// We don't expect any rows in return from a migration.
		if _, err := db.Exec(string(content)); err != nil {
			return err
		}
	}

	return nil
}
