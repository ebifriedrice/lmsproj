package handlers

import (
	"database/sql"
	"html/template"

	"github.com/alexedwards/scs/v2"
)

// Handlers struct holds dependencies for handlers.
type Handlers struct {
	DB             *sql.DB
	SessionManager *scs.SessionManager
	TemplateCache  map[string]*template.Template
}

// NewHandlers creates a new Handlers struct.
func NewHandlers(db *sql.DB, sessionManager *scs.SessionManager) (*Handlers, error) {
	// Initialize a new template cache.
	cache, err := NewTemplateCache()
	if err != nil {
		return nil, err
	}

	return &Handlers{
		DB:             db,
		SessionManager: sessionManager,
		TemplateCache:  cache,
	}, nil
}
