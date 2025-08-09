package handlers

import (
	"fmt"
	"lms/internal/database"
	"lms/internal/models"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// CreateCourseForm displays the form for creating a new course.
func (h *Handlers) CreateCourseForm(w http.ResponseWriter, r *http.Request) {
	td := h.newTemplateData(r)
	h.render(w, r, "admin_create_course.page.tmpl", td)
}

// CreateCourse handles the submission of the new course form.
func (h *Handlers) CreateCourse(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	title := r.PostForm.Get("title")
	description := r.PostForm.Get("description")

	if title == "" {
		// In a real app, you would redirect back with an error.
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	_, err = database.CreateCourse(h.DB, title, description)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Redirect to an admin dashboard or course list page.
	// For now, let's redirect to a placeholder admin page.
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handlers) ShowCourse(w http.ResponseWriter, r *http.Request) {
	// Get the course ID from the URL parameter.
	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	// Fetch the course from the database.
	course, err := database.GetCourse(h.DB, courseID)
	if err != nil {
		// Handle case where course is not found
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	// Fetch the lessons for the course.
	lessons, err := database.GetLessonsForCourse(h.DB, courseID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	td := h.newTemplateData(r)
	td.Data["Course"] = course
	td.Data["Lessons"] = lessons

	h.render(w, r, "admin_course_detail.page.tmpl", td)
}

func (h *Handlers) CreateLesson(w http.ResponseWriter, r *http.Request) {
	// Get the course ID from the URL parameter.
	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	// Parse the form data.
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	title := r.PostForm.Get("title")
	positionStr := r.PostForm.Get("position")

	if title == "" || positionStr == "" {
		http.Error(w, "Title and position are required", http.StatusBadRequest)
		return
	}

	position, err := strconv.Atoi(positionStr)
	if err != nil {
		http.Error(w, "Invalid position", http.StatusBadRequest)
		return
	}

	// Create the lesson in the database.
	_, err = database.CreateLesson(h.DB, courseID, title, position)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Redirect back to the course detail page.
	http.Redirect(w, r, fmt.Sprintf("/admin/courses/%d", courseID), http.StatusSeeOther)
}

func (h *Handlers) ShowLesson(w http.ResponseWriter, r *http.Request) {
	lessonID, err := strconv.ParseInt(chi.URLParam(r, "lessonID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid lesson ID", http.StatusBadRequest)
		return
	}

	// Fetch existing content to display it.
	video, _ := database.GetVideoByLessonID(h.DB, lessonID)
	text, _ := database.GetTextByLessonID(h.DB, lessonID)
	mcq, _ := database.GetMCQByLessonID(h.DB, lessonID)

	// We ignore errors here because content might not exist, and that's okay.
	// The template will handle the nil cases.

	td := h.newTemplateData(r)
	td.Data["LessonID"] = lessonID
	td.Data["Video"] = video
	td.Data["Text"] = text
	td.Data["MCQ"] = mcq

	h.render(w, r, "admin_lesson_detail.page.tmpl", td)
}

func (h *Handlers) AddContent(w http.ResponseWriter, r *http.Request) {
	lessonID, err := strconv.ParseInt(chi.URLParam(r, "lessonID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid lesson ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	contentType := r.PostForm.Get("contentType")

	switch contentType {
	case "video":
		title := r.PostForm.Get("videoTitle")
		url := r.PostForm.Get("videoURL")
		if title == "" || url == "" {
			http.Error(w, "Title and URL are required for video", http.StatusBadRequest)
			return
		}
		// In a real app, you'd handle updates instead of just creating.
		// For now, we'll just create, which will fail if content already exists due to UNIQUE constraint.
		// A better approach would be an "upsert".
		_, err = database.CreateVideo(h.DB, lessonID, title, url)

	case "text":
		title := r.PostForm.Get("textTitle")
		content := r.PostForm.Get("textContent")
		if title == "" || content == "" {
			http.Error(w, "Title and content are required for text", http.StatusBadRequest)
			return
		}
		_, err = database.CreateText(h.DB, lessonID, title, content)

	case "mcq":
		question := r.PostForm.Get("mcqQuestion")
		options := []string{
			r.PostForm.Get("mcqOption0"),
			r.PostForm.Get("mcqOption1"),
			r.PostForm.Get("mcqOption2"),
			r.PostForm.Get("mcqOption3"),
		}
		correctOptionIndex, _ := strconv.Atoi(r.PostForm.Get("correctOption"))

		_, err = database.CreateMCQ(h.DB, lessonID, question, options, correctOptionIndex)

	default:
		http.Error(w, "Invalid content type", http.StatusBadRequest)
		return
	}

	if err != nil {
		// This will catch the UNIQUE constraint error if content already exists.
		// A real app should handle this more gracefully (e.g., "Content already exists, update it?").
		http.Error(w, "Failed to create content: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect back to the lesson detail page.
	http.Redirect(w, r, fmt.Sprintf("/admin/lessons/%d", lessonID), http.StatusSeeOther)
}

func (h *Handlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := database.GetAllUsers(h.DB)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	td := h.newTemplateData(r)
	td.Data["Users"] = users

	h.render(w, r, "admin_users_list.page.tmpl", td)
}

func (h *Handlers) ShowUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get user details
	user, err := database.GetUserByID(h.DB, userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get enrolled courses
	enrolledCourses, err := database.GetEnrolledCoursesForStudent(h.DB, userID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get all courses to populate the enrollment form
	allCourses, err := database.GetAllCourses(h.DB)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Filter out already enrolled courses from the list of all courses
	enrolledMap := make(map[int64]bool)
	for _, course := range enrolledCourses {
		enrolledMap[course.ID] = true
	}
	var availableCourses []*models.Course
	for _, course := range allCourses {
		if !enrolledMap[course.ID] {
			availableCourses = append(availableCourses, course)
		}
	}

	td := h.newTemplateData(r)
	td.Data["User"] = user
	td.Data["EnrolledCourses"] = enrolledCourses
	td.Data["AvailableCourses"] = availableCourses

	h.render(w, r, "admin_user_detail.page.tmpl", td)
}

func (h *Handlers) GenerateCertificate(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	// An admin can override completion status and generate a certificate.
	_, err = database.CreateCertificate(h.DB, userID, courseID)
	if err != nil {
		// This might fail if a certificate already exists (UNIQUE constraint on token).
		// A more robust implementation would handle this, but for now, an error is acceptable.
		http.Error(w, "Failed to generate certificate: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect back to the user detail page.
	http.Redirect(w, r, fmt.Sprintf("/admin/users/%d", userID), http.StatusSeeOther)
}

func (h *Handlers) EnrollUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	courseID, err := strconv.ParseInt(r.PostForm.Get("courseID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	err = database.EnrollStudentInCourse(h.DB, userID, courseID)
	if err != nil {
		http.Error(w, "Failed to enroll user", http.StatusInternalServerError)
		return
	}

	// Redirect back to the user detail page.
	http.Redirect(w, r, fmt.Sprintf("/admin/users/%d", userID), http.StatusSeeOther)
}
