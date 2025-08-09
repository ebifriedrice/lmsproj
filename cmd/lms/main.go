package main

import (
	"database/sql"
	"lms/internal/database"
	"lms/internal/handlers"
	"lms/internal/middleware"
	"log"
	"time"

	"net/http"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

// Application struct holds the dependencies for the Application.
type Application struct {
	db             *sql.DB
	sessionManager *scs.SessionManager
	handlers       *handlers.Handlers
	middleware     *middleware.Middleware
}

func (app *Application) Routes() http.Handler {
	mux := chi.NewRouter()

	// Register global middleware
	mux.Use(chiMiddleware.Logger)
	mux.Use(chiMiddleware.Recoverer)
	mux.Use(app.sessionManager.LoadAndSave)

	// Public routes
	mux.Group(func(r chi.Router) {
		r.Get("/", app.handlers.Dashboard)
		r.Get("/courses/{courseID}", app.handlers.ShowCourse)
		r.Get("/lessons/{lessonID}", app.handlers.ShowLesson)
		r.Get("/register", app.handlers.RegisterForm)
		r.Post("/register", app.handlers.Register)
		r.Get("/login", app.handlers.LoginForm)
		r.Post("/login", app.handlers.Login)
		r.Post("/logout", app.handlers.Logout)
		r.Get("/certificates/{token}", app.handlers.ViewCertificate)

		// Serve static files
		fs := http.FileServer(http.Dir("./web/static/"))
		r.Handle("/static/*", http.StripPrefix("/static/", fs))

	})
	// Protected routes for authenticated users
	mux.Group(func(r chi.Router) {
		r.Use(app.middleware.RequireAuthentication)

		r.Post("/mcqs/{mcqID}/submit", app.handlers.SubmitMCQ)
		r.Post("/lessons/{lessonID}/complete", app.handlers.MarkLessonComplete)
	})

	// Admin routes
	mux.Route("/admin", func(r chi.Router) {
		r.Use(app.middleware.RequireAuthentication)
		r.Use(app.middleware.RequireAdmin)

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Welcome to the Admin Dashboard"))
		})
		r.Get("/courses/new", app.handlers.CreateCourseForm)
		r.Post("/courses/new", app.handlers.CreateCourse)
		r.Get("/courses/{courseID}", app.handlers.ShowCourseAdmin)
		r.Post("/courses/{courseID}/lessons", app.handlers.CreateLesson)
		r.Get("/lessons/{lessonID}", app.handlers.ShowLessonAdmin)
		r.Post("/lessons/{lessonID}/content", app.handlers.AddContent)
		r.Get("/users", app.handlers.ListUsers)
		r.Get("/users/{userID}", app.handlers.ShowUser)
		r.Post("/users/{userID}/enroll", app.handlers.EnrollUser)
		r.Post("/users/{userID}/courses/{courseID}/generate-certificate", app.handlers.GenerateCertificate)
	})

	return mux
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
	app := &Application{
		db:             db,
		sessionManager: sessionManager,
		handlers:       h,
		middleware:     mw,
	}

	// Set up the HTTP server.
	srv := &http.Server{
		Addr:    ":8080",
		Handler: app.Routes(), // Use the chi router
	}

	log.Printf("Starting server on %s", srv.Addr)
	err = srv.ListenAndServe()
	log.Fatal(err)
}
