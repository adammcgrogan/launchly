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

	var general, industry []templateEntry

	for _, t := range siteTemplates {
		entry := templateEntry{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			ExampleURL:  h.exampleURL(t.ExampleSlug),
			Industry:    t.Industry,
		}
		if t.Industry == "" {
			general = append(general, entry)
		} else {
			industry = append(industry, entry)
		}
	}

	tmpl := template.Must(template.ParseFiles(
		"web/templates/public/home_base.html",
		"web/templates/public/home.html",
	))
	tmpl.ExecuteTemplate(w, "base", map[string]any{
		"GeneralTemplates":  general,
		"IndustryTemplates": industry,
	})
}
