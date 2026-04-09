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

	var general []templateEntry
	industryMap := make(map[string][]templateEntry)
	var industryOrder []string

	for _, t := range siteTemplates {
		entry := templateEntry{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			ExampleURL:  h.exampleURL(t.ExampleSlug),
		}
		if t.Industry == "" {
			general = append(general, entry)
		} else {
			if _, seen := industryMap[t.Industry]; !seen {
				industryOrder = append(industryOrder, t.Industry)
			}
			industryMap[t.Industry] = append(industryMap[t.Industry], entry)
		}
	}

	var industryGroups []industryGroup
	for _, name := range industryOrder {
		industryGroups = append(industryGroups, industryGroup{
			Name:      name,
			Templates: industryMap[name],
		})
	}

	tmpl := template.Must(template.ParseFiles(
		"web/templates/public/home_base.html",
		"web/templates/public/home.html",
	))
	tmpl.ExecuteTemplate(w, "base", map[string]any{
		"GeneralTemplates": general,
		"IndustryGroups":   industryGroups,
	})
}
