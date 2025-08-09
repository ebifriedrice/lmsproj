package handlers

import (
	"database/sql"
	"fmt"
	"lms/internal/database"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	// Get user data from the session.
	userID := h.SessionManager.GetInt64(r.Context(), "authenticatedUserID")
	userRole := h.SessionManager.GetString(r.Context(), "userRole")

	// If the user is an admin, maybe redirect to an admin dashboard.
	if userRole == "admin" {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// For students, show their enrolled courses.
	enrolledCourses, err := database.GetEnrolledCoursesForStudent(h.DB, userID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	td := h.newTemplateData(r)
	td.Data["Courses"] = enrolledCourses

	// Render the dashboard template.
	h.render(w, r, "dashboard.page.tmpl", td)
}

func (h *Handlers) ShowCourse(w http.ResponseWriter, r *http.Request) {
	// Get the course ID from the URL parameter.
	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	// TODO: Check if the student is enrolled in this course.
	// For now, we'll allow any authenticated user to see the course page.

	// Fetch the course from the database.
	course, err := database.GetCourse(h.DB, courseID)
	if err != nil {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}

	// Fetch the lessons for the course.
	lessons, err := database.GetLessonsForCourse(h.DB, courseID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	userID := h.SessionManager.GetInt64(r.Context(), "authenticatedUserID")

	// Get completed lessons for the user
	completedLessons, err := database.GetCompletedLessonsForUser(h.DB, userID, courseID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	td := h.newTemplateData(r)
	td.Data["Course"] = course
	td.Data["Lessons"] = lessons
	td.Data["CompletedLessons"] = completedLessons

	h.render(w, r, "course_detail.page.tmpl", td)
}

func (h *Handlers) SubmitMCQ(w http.ResponseWriter, r *http.Request) {
	// Get the MCQ ID from the URL parameter.
	mcqID, err := strconv.ParseInt(chi.URLParam(r, "mcqID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid MCQ ID", http.StatusBadRequest)
		return
	}

	// Get the user ID from the session.
	userID := h.SessionManager.GetInt64(r.Context(), "authenticatedUserID")

	// Parse the form data.
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	selectedOptionStr := r.PostForm.Get("option")
	selectedOption, err := strconv.Atoi(selectedOptionStr)
	if err != nil {
		http.Error(w, "Invalid option", http.StatusBadRequest)
		return
	}

	// Submit the MCQ answer.
	submission, err := database.SubmitMCQ(h.DB, userID, mcqID, selectedOption)
	if err != nil {
		// Could be a unique constraint violation if already submitted.
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get the lesson ID for the redirect.
	mcq, err := database.GetMCQByID(h.DB, mcqID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// For now, just a simple message. A better UX would be to show the result on the lesson page.
	if submission.IsCorrect {
		// You could use sessions to flash a success message.
		// h.SessionManager.Put(r.Context(), "flash", "Correct!")
		http.Redirect(w, r, fmt.Sprintf("/lessons/%d", mcq.LessonID), http.StatusSeeOther)
	} else {
		// h.SessionManager.Put(r.Context(), "flash", "Incorrect. Try again!")
		http.Redirect(w, r, fmt.Sprintf("/lessons/%d", mcq.LessonID), http.StatusSeeOther)
	}
}

func (h *Handlers) ShowLesson(w http.ResponseWriter, r *http.Request) {
	// Get the lesson ID from the URL parameter.
	lessonID, err := strconv.ParseInt(chi.URLParam(r, "lessonID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid lesson ID", http.StatusBadRequest)
		return
	}

	// TODO: Check if the student is enrolled in the course this lesson belongs to.

	// Fetch the content for the lesson.
	// We use Get...ByLessonID functions. They will return sql.ErrNoRows if content doesn't exist.
	video, errVideo := database.GetVideoByLessonID(h.DB, lessonID)
	if errVideo != nil && errVideo != sql.ErrNoRows {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	text, errText := database.GetTextByLessonID(h.DB, lessonID)
	if errText != nil && errText != sql.ErrNoRows {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	mcq, errMCQ := database.GetMCQByLessonID(h.DB, lessonID)
	if errMCQ != nil && errMCQ != sql.ErrNoRows {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	userID := h.SessionManager.GetInt64(r.Context(), "authenticatedUserID")
	isComplete, err := database.IsLessonComplete(h.DB, userID, lessonID)
	if err != nil {
		isComplete = false
	}

	td := h.newTemplateData(r)
	td.Data["LessonID"] = lessonID
	td.Data["Video"] = video
	td.Data["Text"] = text
	td.Data["MCQ"] = mcq
	td.Data["IsComplete"] = isComplete

	h.render(w, r, "lesson_detail.page.tmpl", td)
}

func (h *Handlers) MarkLessonComplete(w http.ResponseWriter, r *http.Request) {
	lessonID, err := strconv.ParseInt(chi.URLParam(r, "lessonID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid lesson ID", http.StatusBadRequest)
		return
	}

	userID := h.SessionManager.GetInt64(r.Context(), "authenticatedUserID")

	err = database.MarkLessonAsComplete(h.DB, userID, lessonID)
	if err != nil {
		// This might fail if the lesson is already marked as complete (UNIQUE constraint)
		// For now, we'll just log it. A more robust solution would handle this gracefully.
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get the lesson to find the course ID for the completion check and redirect.
	lesson, err := database.GetLesson(h.DB, lessonID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Check if the course is now complete.
	isComplete, err := database.IsCourseComplete(h.DB, userID, lesson.CourseID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if isComplete {
		// If the course is complete, generate a certificate.
		// We should probably check if a certificate already exists first.
		_, err := database.CreateCertificate(h.DB, userID, lesson.CourseID)
		if err != nil {
			// This might fail if a certificate already exists. We can ignore this for now.
		}
	}

	// On success, return an HTML snippet to be swapped in.
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<div class="text-green-500 font-bold">✓ Completed</div>`)
}

func (h *Handlers) ViewCertificate(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	details, err := database.GetCertificateDetailsByToken(h.DB, token)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Certificate not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	td := h.newTemplateData(r)
	td.Data["Certificate"] = details
	h.render(w, r, "certificate.page.tmpl", td)
}
