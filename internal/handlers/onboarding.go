package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/adammcgrogan/launchly/internal/models"
)

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func toSlug(s string) string {
	s = strings.ToLower(s)
	s = slugRe.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func (h *Handler) OnboardingForm(w http.ResponseWriter, r *http.Request) {
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

	h.render(w, "onboarding", map[string]any{
		"AllTemplates": all,
	})
}

func (h *Handler) OnboardingSubmit(w http.ResponseWriter, r *http.Request) {
	// Rate limit: 3 submissions per IP per 5 minutes
	ip := clientIP(r)
	if !h.onboardingLimiter.allow(ip) {
		http.Error(w, "Too many requests — please wait a moment and try again.", http.StatusTooManyRequests)
		return
	}

	// Limit request body to prevent abuse
	r.Body = http.MaxBytesReader(w, r.Body, 256*1024)

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Honeypot: bots fill hidden fields, humans don't
	if r.FormValue("website") != "" {
		http.Redirect(w, r, "/get-started?thanks=1", http.StatusSeeOther)
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
		MapEmbedURL:  r.FormValue("map_embed_url"),
		FacebookURL:  r.FormValue("facebook_url"),
		InstagramURL: r.FormValue("instagram_url"),
		WhatsAppURL:  r.FormValue("whatsapp_url"),
		TwitterURL:   r.FormValue("twitter_url"),
		TikTokURL:    r.FormValue("tiktok_url"),
		LinkedInURL:  r.FormValue("linkedin_url"),
		YouTubeURL:   r.FormValue("youtube_url"),
		LeadEmail:    r.FormValue("lead_email"),
		Plan:         r.FormValue("plan"),
		Status:       models.StatusDraft,
	}

	if site.Template == "" {
		site.Template = "bold"
	}

	if err := h.store.CreateSite(site); err != nil {
		http.Error(w, "could not save your submission", http.StatusInternalServerError)
		return
	}

	if site.LeadEmail != "" {
		if err := h.email.SendWelcomeEmail(site.LeadEmail, site.BusinessName); err != nil {
			log.Printf("send welcome email error: %v", err)
		}
	}

	if err := h.email.SendNewSubmissionNotification("hello@launchly.ltd", site.BusinessName, site.Template, site.Location, site.LeadEmail); err != nil {
		log.Printf("send submission notification error: %v", err)
	}

	h.render(w, "thankyou", map[string]any{
		"BusinessName": site.BusinessName,
	})
}
