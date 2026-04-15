package handlers

import (
	"net/http"
)

// buildEntry converts a siteTemplate row to a templateEntry for public pages.
func (h *Handler) buildEntry(t struct {
	ID          string
	Name        string
	Description string
	ExampleSlug string
	Category    string
	Industry    string
	Tags        []string
	Palettes    []Palette
}) templateEntry {
	return templateEntry{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		ExampleURL:  h.exampleURL(t.ExampleSlug),
		Category:    t.Category,
		Industry:    t.Industry,
		Tags:        t.Tags,
	}
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	var general, specific []templateEntry
	for _, t := range siteTemplates {
		e := h.buildEntry(t)
		if t.Category == "general" {
			general = append(general, e)
		} else {
			specific = append(specific, e)
		}
	}

	featured := general

	h.render(w, "home", map[string]any{
		"FeaturedTemplates": featured,
		"TotalTemplates":    len(general) + len(specific),
	})
}

func (h *Handler) TemplatesPage(w http.ResponseWriter, r *http.Request) {
	var general, specific []templateEntry
	for _, t := range siteTemplates {
		e := h.buildEntry(t)
		if t.Category == "general" {
			general = append(general, e)
		} else {
			specific = append(specific, e)
		}
	}

	h.render(w, "templates", map[string]any{
		"GeneralTemplates":  general,
		"SpecificTemplates": specific,
		"TotalTemplates":    len(general) + len(specific),
	})
}
