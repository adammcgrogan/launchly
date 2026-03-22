package handlers

import (
	"html/template"
	"net/http"
)

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	type templateEntry struct {
		ID          string
		Name        string
		Description string
		ExampleSlug string
		ExampleURL  string
	}

	entries := make([]templateEntry, len(siteTemplates))
	for i, t := range siteTemplates {
		entries[i] = templateEntry{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			ExampleSlug: t.ExampleSlug,
			ExampleURL:  h.exampleURL(r.Host, t.ExampleSlug),
		}
	}

	tmpl := template.Must(template.ParseFiles(
		"web/templates/public/home_base.html",
		"web/templates/public/home.html",
	))
	tmpl.ExecuteTemplate(w, "base", map[string]any{
		"Templates": entries,
	})
}
