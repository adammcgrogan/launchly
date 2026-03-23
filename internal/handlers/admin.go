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
	site, err := h.store.GetSiteByID(id)
	if err != nil || site == nil {
		http.NotFound(w, r)
		return
	}
	if err := h.store.PublishSite(id); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	if site.LeadEmail != "" {
		siteURL := h.baseURL(r.Host) + "/sites/" + site.Slug
		if err := h.email.SendSitePublished(site.LeadEmail, site.BusinessName, siteURL); err != nil {
			log.Printf("send site published email error: %v", err)
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/sites/%d", id), http.StatusSeeOther)
}

func (h *Handler) AdminUnpublish(w http.ResponseWriter, r *http.Request) {
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
	if err := h.store.UnpublishSite(id); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	if site.LeadEmail != "" {
		if err := h.email.SendSiteUnpublished(site.LeadEmail, site.BusinessName); err != nil {
			log.Printf("send site unpublished email error: %v", err)
		}
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

func siteDiff(old, updated *models.Site) []string {
	var changes []string
	check := func(label, a, b string) {
		if a != b {
			changes = append(changes, label)
		}
	}
	check("Business Name", old.BusinessName, updated.BusinessName)
	check("Tagline", old.Tagline, updated.Tagline)
	check("About", old.About, updated.About)
	check("Services", old.Services, updated.Services)
	check("Certifications / Trust Badges", old.Certifications, updated.Certifications)
	check("Location", old.Location, updated.Location)
	check("CTA Button Text", old.CTAText, updated.CTAText)
	check("Testimonials", old.Testimonials, updated.Testimonials)
	check("Logo", old.LogoURL, updated.LogoURL)
	check("Photo Gallery", old.Gallery, updated.Gallery)
	check("Phone", old.Phone, updated.Phone)
	check("Business Email", old.Email, updated.Email)
	check("Address", old.Address, updated.Address)
	check("Opening Hours", old.Hours, updated.Hours)
	check("Google Maps URL", old.MapURL, updated.MapURL)
	check("Facebook", old.FacebookURL, updated.FacebookURL)
	check("Instagram", old.InstagramURL, updated.InstagramURL)
	check("WhatsApp", old.WhatsAppURL, updated.WhatsAppURL)
	check("Twitter / X", old.TwitterURL, updated.TwitterURL)
	check("TikTok", old.TikTokURL, updated.TikTokURL)
	check("LinkedIn", old.LinkedInURL, updated.LinkedInURL)
	check("YouTube", old.YouTubeURL, updated.YouTubeURL)
	return changes
}

func (h *Handler) AdminUpdateSite(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	old, err := h.store.GetSiteByID(id)
	if err != nil || old == nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	// Copy old into updated, then apply form values
	updated := *old
	updated.BusinessName = strings.TrimSpace(r.FormValue("business_name"))
	updated.Tagline = strings.TrimSpace(r.FormValue("tagline"))
	updated.About = strings.TrimSpace(r.FormValue("about"))
	updated.Services = strings.TrimSpace(r.FormValue("services"))
	updated.Certifications = strings.TrimSpace(r.FormValue("certifications"))
	updated.Location = strings.TrimSpace(r.FormValue("location"))
	updated.CTAText = strings.TrimSpace(r.FormValue("cta_text"))
	updated.Testimonials = buildTestimonials(r)
	updated.LogoURL = strings.TrimSpace(r.FormValue("logo_url"))
	updated.Gallery = strings.TrimSpace(r.FormValue("gallery"))
	updated.Phone = strings.TrimSpace(r.FormValue("phone"))
	updated.Email = strings.TrimSpace(r.FormValue("email"))
	updated.Address = strings.TrimSpace(r.FormValue("address"))
	updated.Hours = strings.TrimSpace(r.FormValue("hours"))
	updated.MapURL = strings.TrimSpace(r.FormValue("map_url"))
	updated.FacebookURL = strings.TrimSpace(r.FormValue("facebook_url"))
	updated.InstagramURL = strings.TrimSpace(r.FormValue("instagram_url"))
	updated.WhatsAppURL = strings.TrimSpace(r.FormValue("whatsapp_url"))
	updated.TwitterURL = strings.TrimSpace(r.FormValue("twitter_url"))
	updated.TikTokURL = strings.TrimSpace(r.FormValue("tiktok_url"))
	updated.LinkedInURL = strings.TrimSpace(r.FormValue("linkedin_url"))
	updated.YouTubeURL = strings.TrimSpace(r.FormValue("youtube_url"))
	updated.UmamiWebsiteID = strings.TrimSpace(r.FormValue("umami_website_id"))
	updated.LeadEmail = strings.TrimSpace(r.FormValue("lead_email"))
	changes := siteDiff(old, &updated)
	if err := h.store.UpdateSite(&updated); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	if len(changes) > 0 && updated.LeadEmail != "" {
		if err := h.email.SendSiteUpdated(updated.LeadEmail, updated.BusinessName, changes); err != nil {
			log.Printf("send site updated email error: %v", err)
		}
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
	if err != nil || site == nil {
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

func (h *Handler) AdminCancelSubscription(w http.ResponseWriter, r *http.Request) {
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
	if site.StripeSubscriptionID == "" {
		http.Error(w, "no subscription on record", http.StatusBadRequest)
		return
	}
	if err := h.pay.CancelSubscription(site.StripeSubscriptionID); err != nil {
		http.Error(w, "stripe error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.store.SetSiteCancelled(site.StripeSubscriptionID); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/sites/%d", id), http.StatusSeeOther)
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
	switch event.Type {
	case "checkout.session.completed":
		if event.SessionID != "" {
			if err := h.store.SetSitePaid(event.SessionID, event.SubscriptionID); err != nil {
				log.Printf("set site paid error: %v", err)
			} else {
				log.Printf("payment received for session %s", event.SessionID)
			}
		}
	case "customer.subscription.deleted":
		if event.SubscriptionID != "" {
			if err := h.store.SetSiteCancelled(event.SubscriptionID); err != nil {
				log.Printf("set site cancelled error: %v", err)
			} else {
				log.Printf("subscription cancelled: %s", event.SubscriptionID)
			}
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
	mux.HandleFunc("POST /admin/sites/{id}/cancel-subscription", h.adminAuth(h.AdminCancelSubscription))
}
