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

	var all []templateEntry
	for _, t := range siteTemplates {
		all = append(all, templateEntry{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			ExampleURL:  h.exampleURL(t.ExampleSlug),
			Industry:    t.Industry,
			Tags:        t.Tags,
		})
	}

	featured := all
	if len(featured) > 8 {
		featured = featured[:8]
	}

	tmpl := template.Must(template.ParseFiles(
		"web/templates/public/home_base.html",
		"web/templates/public/home.html",
	))
	tmpl.ExecuteTemplate(w, "base", map[string]any{
		"FeaturedTemplates": featured,
		"TotalTemplates":    len(all),
	})
}

func (h *Handler) TemplatesPage(w http.ResponseWriter, r *http.Request) {
	var all []templateEntry
	for _, t := range siteTemplates {
		all = append(all, templateEntry{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			ExampleURL:  h.exampleURL(t.ExampleSlug),
			Industry:    t.Industry,
			Tags:        t.Tags,
		})
	}

	tmpl := template.Must(template.ParseFiles(
		"web/templates/public/home_base.html",
		"web/templates/public/templates.html",
	))
	tmpl.ExecuteTemplate(w, "base", map[string]any{
		"AllTemplates":   all,
		"TotalTemplates": len(all),
	})
}
