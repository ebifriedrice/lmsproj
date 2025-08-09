package main

import (
	"database/sql"
	"log"
	"lms/internal/database"
	"lms/internal/handlers"
	"lms/internal/middleware"
	"net/http"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
)

// application struct holds the dependencies for the application.
type application struct {
	db             *sql.DB
	sessionManager *scs.SessionManager
	handlers       *handlers.Handlers
	middleware     *middleware.Middleware
}

func main() {
	// Define the Data Source Name (DSN) for the SQLite database.
	const dsn = "lms.db"

	// Establish a connection to the database.
	db, err := database.NewDB(dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Successfully connected to the database.")

	// Apply database migrations.
	if err := database.ApplyMigrations(db, "migrations"); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	log.Println("Database migrations applied successfully.")

	// Initialize a new session manager and configure it.
	sessionManager := scs.New()
	sessionManager.Store = sqlite3store.New(db)
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.IdleTimeout = 20 * time.Minute
	sessionManager.Cookie.Persist = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = false

	// Create a new Handlers struct, which now includes the template cache.
	h, err := handlers.NewHandlers(db, sessionManager)
	if err != nil {
		log.Fatalf("failed to create handlers: %v", err)
	}

	// Create a new Middleware struct.
	mw := middleware.NewMiddleware(sessionManager)

	// Create an instance of the application struct.
	app := &application{
		db:             db,
		sessionManager: sessionManager,
		handlers:       h,
		middleware:     mw,
	}

	// Set up the HTTP server.
	srv := &http.Server{
		Addr:    ":8080",
		Handler: app.routes(), // Use the chi router
	}

	log.Printf("Starting server on %s", srv.Addr)
	err = srv.ListenAndServe()
	log.Fatal(err)
}
