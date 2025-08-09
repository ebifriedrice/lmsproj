package handlers

import (
	"fmt"
	"lms/internal/database"
	"net/http"
)

func (h *Handlers) RegisterForm(w http.ResponseWriter, r *http.Request) {
	td := h.newTemplateData(r)
	h.render(w, r, "register.page.tmpl", td)
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	// Basic validation
	if username == "" || password == "" {
		// In a real app, you'd redirect back with an error message.
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// For now, all new users are students.
	_, err = database.CreateUser(h.DB, username, password, "student")
	if err != nil {
		// This could be a unique constraint violation if the username is taken.
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Redirect to the login page after successful registration.
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handlers) LoginForm(w http.ResponseWriter, r *http.Request) {
	td := h.newTemplateData(r)
	h.render(w, r, "login.page.tmpl", td)
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	user, err := database.AuthenticateUser(h.DB, username, password)
	if err != nil {
		// If authentication fails, redirect back to the login page.
		// You might want to show an error message.
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Authentication successful. Store the user ID and role in the session.
	h.SessionManager.Put(r.Context(), "authenticatedUserID", user.ID)
	h.SessionManager.Put(r.Context(), "userRole", user.Role)

	// Redirect to a dashboard page. For now, let's redirect to "/".
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	// Destroy the session data.
	err := h.SessionManager.Destroy(r.Context())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// Redirect to the login page.
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
