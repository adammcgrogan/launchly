package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"

	"github.com/adammcgrogan/locallaunch/internal/models"
)

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func toSlug(s string) string {
	s = strings.ToLower(s)
	s = slugRe.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func (h *Handler) OnboardingForm(w http.ResponseWriter, r *http.Request) {
	type templateEntry struct {
		ID          string
		Name        string
		Description string
		ExampleURL  string
	}
	entries := make([]templateEntry, len(siteTemplates))
	for i, t := range siteTemplates {
		entries[i] = templateEntry{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			ExampleURL:  h.exampleURL(r.Host, t.ExampleSlug),
		}
	}

	tmpl := template.Must(template.ParseFiles(
		"web/templates/public/home_base.html",
		"web/templates/public/onboarding.html",
	))
	tmpl.ExecuteTemplate(w, "base", map[string]any{
		"Templates": entries,
	})
}

func (h *Handler) OnboardingSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	businessName := strings.TrimSpace(r.FormValue("business_name"))
	if businessName == "" {
		http.Error(w, "business name is required", http.StatusBadRequest)
		return
	}

	slug := toSlug(businessName)
	base := slug
	for i := 2; ; i++ {
		existing, _ := h.store.GetSiteBySlug(slug)
		if existing == nil {
			break
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}

	site := &models.Site{
		Slug:         slug,
		BusinessName: businessName,
		Template:     r.FormValue("template"),
		Tagline:      r.FormValue("tagline"),
		About:        r.FormValue("about"),
		Services:       r.FormValue("services"),
		Location:       strings.TrimSpace(r.FormValue("location")),
		Certifications: strings.TrimSpace(r.FormValue("certifications")),
		CTAText:        strings.TrimSpace(r.FormValue("cta_text")),
		Testimonials: buildTestimonials(r),
		LogoURL:      r.FormValue("logo_url"),
		Gallery:      r.FormValue("gallery"),
		Phone:        r.FormValue("phone"),
		Email:        r.FormValue("email"),
		Address:      r.FormValue("address"),
		Hours:        r.FormValue("hours"),
		MapURL:       r.FormValue("map_url"),
		FacebookURL:  r.FormValue("facebook_url"),
		InstagramURL: r.FormValue("instagram_url"),
		WhatsAppURL:  r.FormValue("whatsapp_url"),
		TwitterURL:   r.FormValue("twitter_url"),
		TikTokURL:    r.FormValue("tiktok_url"),
		LinkedInURL:  r.FormValue("linkedin_url"),
		YouTubeURL:   r.FormValue("youtube_url"),
		LeadEmail:    r.FormValue("lead_email"),
		Status:       models.StatusDraft,
	}

	if site.Template == "" {
		site.Template = "bold"
	}

	if err := h.store.CreateSite(site); err != nil {
		http.Error(w, "could not save your submission", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles(
		"web/templates/public/home_base.html",
		"web/templates/public/thankyou.html",
	))
	tmpl.ExecuteTemplate(w, "base", map[string]any{
		"BusinessName": site.BusinessName,
	})
}
