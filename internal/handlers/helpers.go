package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

func NewTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// Get all the page templates
	pages, err := filepath.Glob("./web/templates/*.page.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		// Create a new template set for each page.
		// The name of the template set is the page's filename.
		ts, err := template.New(name).ParseFiles("./web/templates/base.layout.tmpl")
		if err != nil {
			return nil, err
		}

		// Parse any component templates (partials)
		ts, err = ts.ParseGlob("./web/templates/components/*.tmpl")
		if err != nil {
			// It's okay if there are no components yet.
			if _, ok := err.(*filepath.GlobError); !ok {
				// return nil, err
			}
		}

		// Parse the page template itself
		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

// TemplateData holds data passed to templates.
type TemplateData struct {
	IsAuthenticated bool
	UserRole        string
	Data            map[string]interface{}
}

// newTemplateData creates a new TemplateData struct with default values.
func (h *Handlers) newTemplateData(r *http.Request) *TemplateData {
	return &TemplateData{
		IsAuthenticated: h.SessionManager.Exists(r.Context(), "authenticatedUserID"),
		UserRole:        h.SessionManager.GetString(r.Context(), "userRole"),
		Data:            make(map[string]interface{}),
	}
}

// render renders a template from the cache.
func (h *Handlers) render(w http.ResponseWriter, r *http.Request, name string, td *TemplateData) {
	ts, ok := h.TemplateCache[name]
	if !ok {
		http.Error(w, fmt.Sprintf("The template %s does not exist", name), http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)

	// Execute the "base" template defined in our layout file.
	err := ts.ExecuteTemplate(buf, "base", td)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
}
