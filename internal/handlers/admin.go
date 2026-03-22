package handlers

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/adammcgrogan/locallaunch/internal/models"
)

func (h *Handler) adminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, pass, ok := r.BasicAuth()
		if !ok || pass != h.adminPass {
			w.Header().Set("WWW-Authenticate", `Basic realm="LocalLaunch Admin"`)
			http.Error(w, "unauthorised", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (h *Handler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	sites, err := h.store.ListSites()
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	tmpl := template.Must(template.ParseFiles(
		"web/templates/admin/base.html",
		"web/templates/admin/dashboard.html",
	))
	tmpl.ExecuteTemplate(w, "base", map[string]any{
		"Sites":   sites,
		"Domain":  h.domain,
		"BaseURL": h.baseURL(r.Host),
	})
}

func (h *Handler) AdminSite(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	site, err := h.store.GetSiteByID(id)
	if err != nil || site == nil {
		http.NotFound(w, r)
		return
	}

	leads, err := h.store.ListLeadsBySite(id)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles(
		"web/templates/admin/base.html",
		"web/templates/admin/site.html",
	))
	tmpl.ExecuteTemplate(w, "base", map[string]any{
		"Site":        site,
		"Leads":       leads,
		"Domain":      h.domain,
		"SiteURL":     h.baseURL(r.Host) + "/sites/" + site.Slug,
		"PaymentSent": r.URL.Query().Get("payment") == "sent",
	})
}

func (h *Handler) AdminPublish(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := h.store.PublishSite(id); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/sites/%d", id), http.StatusSeeOther)
}

func (h *Handler) AdminUnpublish(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := h.store.UnpublishSite(id); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/sites/%d", id), http.StatusSeeOther)
}

func (h *Handler) AdminDeleteSite(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := h.store.DeleteSite(id); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) AdminEditSite(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	site, err := h.store.GetSiteByID(id)
	if err != nil || site == nil {
		http.NotFound(w, r)
		return
	}
	tmpl := template.Must(template.ParseFiles(
		"web/templates/admin/base.html",
		"web/templates/admin/edit.html",
	))
	tmpl.ExecuteTemplate(w, "base", map[string]any{
		"Site":         site,
		"Testimonials": parseTestimonials(site.Testimonials),
	})
}

func (h *Handler) AdminUpdateSite(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	site, err := h.store.GetSiteByID(id)
	if err != nil || site == nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	site.BusinessName = strings.TrimSpace(r.FormValue("business_name"))
	site.Tagline = strings.TrimSpace(r.FormValue("tagline"))
	site.About = strings.TrimSpace(r.FormValue("about"))
	site.Services = strings.TrimSpace(r.FormValue("services"))
	site.Certifications = strings.TrimSpace(r.FormValue("certifications"))
	site.Location = strings.TrimSpace(r.FormValue("location"))
	site.CTAText = strings.TrimSpace(r.FormValue("cta_text"))
	site.Testimonials = buildTestimonials(r)
	site.LogoURL = strings.TrimSpace(r.FormValue("logo_url"))
	site.Gallery = strings.TrimSpace(r.FormValue("gallery"))
	site.Phone = strings.TrimSpace(r.FormValue("phone"))
	site.Email = strings.TrimSpace(r.FormValue("email"))
	site.Address = strings.TrimSpace(r.FormValue("address"))
	site.Hours = strings.TrimSpace(r.FormValue("hours"))
	site.MapURL = strings.TrimSpace(r.FormValue("map_url"))
	site.FacebookURL = strings.TrimSpace(r.FormValue("facebook_url"))
	site.InstagramURL = strings.TrimSpace(r.FormValue("instagram_url"))
	site.WhatsAppURL = strings.TrimSpace(r.FormValue("whatsapp_url"))
	site.TwitterURL = strings.TrimSpace(r.FormValue("twitter_url"))
	site.TikTokURL = strings.TrimSpace(r.FormValue("tiktok_url"))
	site.LinkedInURL = strings.TrimSpace(r.FormValue("linkedin_url"))
	site.YouTubeURL = strings.TrimSpace(r.FormValue("youtube_url"))
	site.UmamiWebsiteID = strings.TrimSpace(r.FormValue("umami_website_id"))
	site.LeadEmail = strings.TrimSpace(r.FormValue("lead_email"))
	if err := h.store.UpdateSite(site); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/sites/%d", id), http.StatusSeeOther)
}

func (h *Handler) AdminSwitchTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	site, err := h.store.GetSiteByID(id)
	if err != nil || site == nil {
		http.NotFound(w, r)
		return
	}
	type templateEntry struct {
		ID          string
		Name        string
		Description string
		ExampleURL  string
		Current     bool
	}
	entries := make([]templateEntry, len(siteTemplates))
	for i, t := range siteTemplates {
		entries[i] = templateEntry{
			ID:          t.ID,
			Name:        t.Name,
			Description: t.Description,
			ExampleURL:  h.exampleURL(r.Host, t.ExampleSlug),
			Current:     t.ID == site.Template,
		}
	}
	tmpl := template.Must(template.ParseFiles(
		"web/templates/admin/base.html",
		"web/templates/admin/switch_template.html",
	))
	tmpl.ExecuteTemplate(w, "base", map[string]any{
		"Site":      site,
		"Templates": entries,
	})
}

func (h *Handler) AdminDoSwitchTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	newTemplate := r.FormValue("template")
	valid := false
	for _, t := range siteTemplates {
		if t.ID == newTemplate {
			valid = true
			break
		}
	}
	if !valid {
		http.Error(w, "invalid template", http.StatusBadRequest)
		return
	}
	if err := h.store.UpdateSiteTemplate(id, newTemplate); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/sites/%d", id), http.StatusSeeOther)
}

func (h *Handler) AdminSendPayment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	site, err := h.store.GetSiteByID(id)
	if err != nil || site == nil || site.Status != models.StatusLive {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	plan := r.FormValue("plan")
	if plan != "starter" && plan != "pro" {
		http.Error(w, "invalid plan", http.StatusBadRequest)
		return
	}
	successURL := h.baseURL(r.Host) + "/payment/success"
	cancelURL := h.baseURL(r.Host) + "/sites/" + site.Slug

	sessionID, checkoutURL, err := h.pay.CreateCheckoutSession(plan, site.LeadEmail, successURL, cancelURL)
	if err != nil {
		http.Error(w, "payment error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.store.SetSitePending(id, plan, sessionID); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	if err := h.email.SendPaymentLink(site.LeadEmail, site.BusinessName, checkoutURL); err != nil {
		log.Printf("send payment link email error: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/sites/%d?payment=sent", id), http.StatusSeeOther)
}

func (h *Handler) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 65536))
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}
	event, err := h.pay.ParseWebhook(body, r.Header.Get("Stripe-Signature"))
	if err != nil {
		log.Printf("stripe webhook error: %v", err)
		http.Error(w, "invalid webhook", http.StatusBadRequest)
		return
	}
	if event.Type == "checkout.session.completed" && event.SessionID != "" {
		if err := h.store.SetSitePaid(event.SessionID); err != nil {
			log.Printf("set site paid error: %v", err)
		} else {
			log.Printf("payment received for session %s", event.SessionID)
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) PaymentSuccess(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"web/templates/public/home_base.html",
		"web/templates/public/payment_success.html",
	))
	tmpl.ExecuteTemplate(w, "base", nil)
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Public
	mux.HandleFunc("GET /", h.Home)
	mux.HandleFunc("GET /get-started", h.OnboardingForm)
	mux.HandleFunc("POST /get-started", h.OnboardingSubmit)
	mux.HandleFunc("GET /payment/success", h.PaymentSuccess)
	mux.HandleFunc("POST /webhooks/stripe", h.StripeWebhook)

	// Path-based site routing (works without wildcard subdomain)
	mux.HandleFunc("GET /sites/{slug}", h.ServeSitePath)
	mux.HandleFunc("POST /sites/{slug}/contact", h.SubmitLeadPath)

	// Admin (basic auth protected)
	mux.HandleFunc("GET /admin", h.adminAuth(h.AdminDashboard))
	mux.HandleFunc("GET /admin/sites/{id}", h.adminAuth(h.AdminSite))
	mux.HandleFunc("GET /admin/sites/{id}/edit", h.adminAuth(h.AdminEditSite))
	mux.HandleFunc("POST /admin/sites/{id}/edit", h.adminAuth(h.AdminUpdateSite))
	mux.HandleFunc("POST /admin/sites/{id}/publish", h.adminAuth(h.AdminPublish))
	mux.HandleFunc("POST /admin/sites/{id}/unpublish", h.adminAuth(h.AdminUnpublish))
	mux.HandleFunc("POST /admin/sites/{id}/delete", h.adminAuth(h.AdminDeleteSite))
	mux.HandleFunc("GET /admin/sites/{id}/switch-template", h.adminAuth(h.AdminSwitchTemplate))
	mux.HandleFunc("POST /admin/sites/{id}/switch-template", h.adminAuth(h.AdminDoSwitchTemplate))
	mux.HandleFunc("POST /admin/sites/{id}/send-payment", h.adminAuth(h.AdminSendPayment))
}
