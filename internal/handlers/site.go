package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/adammcgrogan/locallaunch/internal/models"
)

// ServeSite handles subdomain and custom-domain requests.
func (h *Handler) ServeSite(w http.ResponseWriter, r *http.Request) {
	site, err := h.resolveSite(r)
	if err != nil || site == nil || site.Status != models.StatusLive {
		http.NotFound(w, r)
		return
	}
	h.renderSite(w, r, site, "/contact")
}

// ServeSitePath handles path-based requests (/sites/{slug}) — works everywhere including local dev.
func (h *Handler) ServeSitePath(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	site, err := h.store.GetSiteBySlug(slug)
	if err != nil || site == nil || site.Status != models.StatusLive {
		http.NotFound(w, r)
		return
	}
	h.renderSite(w, r, site, "/sites/"+slug+"/contact")
}

// renderSite renders a site template for the given site.
func (h *Handler) renderSite(w http.ResponseWriter, r *http.Request, site *models.Site, formAction string) {
	tmplFile := "web/templates/sites/" + site.Template + ".html"
	tmpl, err := template.ParseFiles("web/templates/sites/base.html", tmplFile)
	if err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := h.siteData(site, r.URL.Query().Get("lead") == "1", formAction)
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("template render error for site %s: %v", site.Slug, err)
	}
}

// SubmitLead handles the contact form POST on subdomain/custom-domain routed sites.
func (h *Handler) SubmitLead(w http.ResponseWriter, r *http.Request) {
	site, err := h.resolveSite(r)
	if err != nil || site == nil {
		http.NotFound(w, r)
		return
	}
	h.saveLead(w, r, site, "/?lead=1")
}

// SubmitLeadPath handles the contact form POST on path-routed sites.
func (h *Handler) SubmitLeadPath(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	site, err := h.store.GetSiteBySlug(slug)
	if err != nil || site == nil {
		http.NotFound(w, r)
		return
	}
	h.saveLead(w, r, site, "/sites/"+slug+"?lead=1")
}

// saveLead saves a contact form submission for the given site.
func (h *Handler) saveLead(w http.ResponseWriter, r *http.Request, site *models.Site, redirectURL string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Honeypot: silently succeed so bots don't know they were rejected
	if r.FormValue("website") != "" {
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	if site.Status != models.StatusLive {
		http.NotFound(w, r)
		return
	}

	lead := &models.Lead{
		SiteID:  site.ID,
		Name:    strings.TrimSpace(r.FormValue("name")),
		Email:   strings.TrimSpace(r.FormValue("email")),
		Phone:   strings.TrimSpace(r.FormValue("phone")),
		Message: strings.TrimSpace(r.FormValue("message")),
	}

	if lead.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if err := h.store.CreateLead(lead); err != nil {
		http.Error(w, "could not save lead", http.StatusInternalServerError)
		return
	}

	if site.LeadEmail != "" {
		h.email.SendLeadNotification(site.LeadEmail, site.BusinessName, lead.Name, lead.Email, lead.Phone, lead.Message)
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

// resolveSite finds the site for the current request: tries subdomain slug first,
// then falls back to custom domain lookup.
func (h *Handler) resolveSite(r *http.Request) (*models.Site, error) {
	if slug := extractSlug(r, h.domain); slug != "" {
		return h.store.GetSiteBySlug(slug)
	}
	host := effectiveHost(r)
	site, err := h.store.GetSiteByCustomDomain(host)
	if err != nil || site == nil {
		return nil, err
	}
	// Custom domain serving is a Pro feature
	if site.Plan != "pro" || site.PaymentStatus != "paid" {
		return nil, nil
	}
	return site, nil
}

// effectiveHost returns the cleaned hostname for the current request.
func effectiveHost(r *http.Request) string {
	host := r.Header.Get("X-Real-Host")
	if host == "" {
		host = r.Host
	}
	return strings.ToLower(strings.Split(host, ":")[0])
}

// extractSlug pulls the subdomain from the request host.
// e.g. "adam-barbers.launchly.ltd" → "adam-barbers"
func extractSlug(r *http.Request, domain string) string {
	host := effectiveHost(r)
	suffix := "." + domain
	if strings.HasSuffix(host, suffix) {
		return strings.TrimSuffix(host, suffix)
	}
	return ""
}

type Testimonial struct {
	Name  string
	Role  string
	Quote string
}

type templateData struct {
	Site           *models.Site
	Services       []string
	Hours          []string
	Certifications []string
	Testimonials   []Testimonial
	Gallery        []string
	CTAText        string
	LeadSent       bool
	FormAction     string
	UmamiScriptURL string
}

func (h *Handler) siteData(site *models.Site, leadSent bool, formAction string) templateData {
	ctaText := site.CTAText
	if ctaText == "" {
		ctaText = "Get in Touch"
	}
	return templateData{
		Site:           site,
		Services:       splitLines(site.Services),
		Hours:          splitLines(site.Hours),
		Certifications: splitLines(site.Certifications),
		Testimonials:   parseTestimonials(site.Testimonials),
		Gallery:        splitLines(site.Gallery),
		CTAText:        ctaText,
		LeadSent:       leadSent,
		FormAction:     formAction,
		UmamiScriptURL: h.umamiScriptURL,
	}
}

func splitLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

// parseTestimonials parses "Name|Role|Quote" lines (role is optional).
func parseTestimonials(s string) []Testimonial {
	var out []Testimonial
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		t := Testimonial{Name: strings.TrimSpace(parts[0])}
		if len(parts) == 2 {
			t.Quote = strings.TrimSpace(parts[1])
		} else if len(parts) >= 3 {
			t.Role = strings.TrimSpace(parts[1])
			t.Quote = strings.TrimSpace(parts[2])
		}
		if t.Quote != "" {
			out = append(out, t)
		}
	}
	return out
}
