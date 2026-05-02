package handlers

import "net/http"

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Health check (no auth — used by Railway)
	mux.HandleFunc("GET /healthz", h.HealthCheck)

	// Public
	mux.HandleFunc("GET /", h.Home)
	mux.HandleFunc("GET /robots.txt", h.RobotsTxt)
	mux.HandleFunc("GET /sitemap.xml", h.Sitemap)
	mux.HandleFunc("GET /privacy", h.Privacy)
	mux.HandleFunc("GET /terms", h.Terms)
	mux.HandleFunc("GET /templates", h.TemplatesPage)
	mux.HandleFunc("GET /get-started", h.OnboardingForm)
	mux.HandleFunc("POST /get-started", h.OnboardingSubmit)
	mux.HandleFunc("GET /check-slug", h.CheckSlug)
	mux.HandleFunc("GET /payment/success", h.PaymentSuccess)
	mux.HandleFunc("POST /webhooks/stripe", h.StripeWebhook)

	// Path-based site routing (works without wildcard subdomain, useful for local dev)
	mux.HandleFunc("GET /sites/{slug}", h.ServeSitePath)
	mux.HandleFunc("POST /sites/{slug}/contact", h.SubmitLeadPath)

	// Admin login (no auth)
	mux.HandleFunc("GET /admin/login", h.AdminLogin)
	mux.HandleFunc("POST /admin/login", h.AdminLoginPost)
	mux.HandleFunc("GET /admin/logout", h.AdminLogout)

	// Admin (session auth protected)
	mux.HandleFunc("GET /admin", h.adminAuth(h.AdminDashboard))
	mux.HandleFunc("GET /admin/sites/{id}", h.adminAuth(h.AdminSite))
	mux.HandleFunc("GET /admin/sites/{id}/edit", h.adminAuth(h.AdminEditSite))
	mux.HandleFunc("POST /admin/sites/{id}/edit", h.adminAuth(h.AdminUpdateSite))
	mux.HandleFunc("POST /admin/sites/{id}/publish", h.adminAuth(h.AdminPublish))
	mux.HandleFunc("POST /admin/sites/{id}/unpublish", h.adminAuth(h.AdminUnpublish))
	mux.HandleFunc("POST /admin/sites/{id}/delete", h.adminAuth(h.AdminDeleteSite))
	mux.HandleFunc("GET /admin/sites/{id}/switch-template", h.adminAuth(h.AdminSwitchTemplate))
	mux.HandleFunc("POST /admin/sites/{id}/switch-template", h.adminAuth(h.AdminDoSwitchTemplate))
	mux.HandleFunc("POST /admin/sites/{id}/appearance", h.adminAuth(h.AdminUpdateAppearance))
	mux.HandleFunc("POST /admin/sites/{id}/send-payment", h.adminAuth(h.AdminSendPayment))
	mux.HandleFunc("POST /admin/sites/{id}/cancel-subscription", h.adminAuth(h.AdminCancelSubscription))
	mux.HandleFunc("GET /admin/sites/{id}/leads.csv", h.adminAuth(h.AdminExportLeads))
	mux.HandleFunc("POST /admin/sites/{id}/notes", h.adminAuth(h.AdminUpdateNotes))
	mux.HandleFunc("POST /admin/sites/{id}/custom-domain", h.adminAuth(h.AdminSetCustomDomain))
	mux.HandleFunc("GET /admin/sites/{id}/check-domain", h.adminAuth(h.AdminCheckDomain))
	mux.HandleFunc("POST /admin/sites/{id}/analytics-frequency", h.adminAuth(h.AdminUpdateAnalyticsFrequency))
	mux.HandleFunc("POST /admin/sites/{id}/send-analytics", h.adminAuth(h.AdminSendAnalytics))
}
