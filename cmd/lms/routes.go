package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	// Serve static files
	fs := http.FileServer(http.Dir("./web/static/"))
	mux.Handle("/static/*", http.StripPrefix("/static/", fs))

	// Register global middleware
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(app.sessionManager.LoadAndSave)

	// Public routes
	mux.Get("/register", app.handlers.RegisterForm)
	mux.Post("/register", app.handlers.Register)
	mux.Get("/login", app.handlers.LoginForm)
	mux.Post("/login", app.handlers.Login)
	mux.Post("/logout", app.handlers.Logout)
	mux.Get("/certificates/{token}", app.handlers.ViewCertificate)

	// Protected routes for authenticated users
	mux.Group(func(r chi.Router) {
		r.Use(app.middleware.RequireAuthentication)

		r.Get("/", app.handlers.Dashboard)
		r.Get("/courses/{courseID}", app.handlers.ShowCourse)
		r.Get("/lessons/{lessonID}", app.handlers.ShowLesson)
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
		r.Get("/courses/{courseID}", app.handlers.ShowCourse)
		r.Post("/courses/{courseID}/lessons", app.handlers.CreateLesson)
		r.Get("/lessons/{lessonID}", app.handlers.ShowLesson)
		r.Post("/lessons/{lessonID}/content", app.handlers.AddContent)
		r.Get("/users", app.handlers.ListUsers)
		r.Get("/users/{userID}", app.handlers.ShowUser)
		r.Post("/users/{userID}/enroll", app.handlers.EnrollUser)
		r.Post("/users/{userID}/courses/{courseID}/generate-certificate", app.handlers.GenerateCertificate)
	})

	return mux
}
