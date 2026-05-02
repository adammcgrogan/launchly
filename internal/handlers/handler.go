package handlers

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/adammcgrogan/launchly/internal/db"
	"github.com/adammcgrogan/launchly/internal/email"
	"github.com/adammcgrogan/launchly/internal/payment"
)

type Handler struct {
	store             *db.Store
	email             *email.Client
	pay               *payment.Client
	domain            string
	adminPass         string
	umamiScriptURL    string
	tmpl              map[string]*template.Template
	contactLimiter    *rateLimiter
	onboardingLimiter *rateLimiter
}

func New(store *db.Store, email *email.Client, pay *payment.Client, domain, adminPass, umamiScriptURL string) (*Handler, error) {
	h := &Handler{
		store:             store,
		email:             email,
		pay:               pay,
		domain:            domain,
		adminPass:         adminPass,
		umamiScriptURL:    umamiScriptURL,
		tmpl:              make(map[string]*template.Template),
		contactLimiter:    newRateLimiter(1, 10*time.Minute),
		onboardingLimiter: newRateLimiter(1, 10*time.Minute),
	}
	if err := h.loadTemplates(); err != nil {
		return nil, fmt.Errorf("load templates: %w", err)
	}
	return h, nil
}

// loadTemplates parses all templates once at startup and caches them.
func (h *Handler) loadTemplates() error {
	pub := func(key, page string) error {
		t, err := template.ParseFiles(
			"web/templates/public/home_base.html",
			"web/templates/public/"+page+".html",
		)
		if err != nil {
			return fmt.Errorf("public/%s: %w", page, err)
		}
		h.tmpl[key] = t
		return nil
	}
	adm := func(key, page string) error {
		t, err := template.ParseFiles(
			"web/templates/admin/base.html",
			"web/templates/admin/"+page+".html",
		)
		if err != nil {
			return fmt.Errorf("admin/%s: %w", page, err)
		}
		h.tmpl["admin:"+key] = t
		return nil
	}

	for _, p := range []string{"home", "templates", "onboarding", "thankyou", "payment_success", "privacy", "terms"} {
		if err := pub(p, p); err != nil {
			return err
		}
	}
	for _, p := range []string{"dashboard", "site", "edit", "switch_template"} {
		if err := adm(p, p); err != nil {
			return err
		}
	}
	lt, err := template.ParseFiles("web/templates/admin/login.html")
	if err != nil {
		return fmt.Errorf("admin/login: %w", err)
	}
	h.tmpl["admin:login"] = lt

	for _, st := range siteTemplates {
		t, err := template.ParseFiles(
			"web/templates/sites/base.html",
			"web/templates/sites/"+st.ID+".html",
		)
		if err != nil {
			return fmt.Errorf("site/%s: %w", st.ID, err)
		}
		h.tmpl["site:"+st.ID] = t
	}
	return nil
}

// render executes a pre-parsed template by key, writing the result to w.
func (h *Handler) render(w http.ResponseWriter, key string, data any) {
	tmpl, ok := h.tmpl[key]
	if !ok {
		http.Error(w, "template not found", http.StatusInternalServerError)
		slog.Error("render: unknown template key", "key", key)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		slog.Error("template render failed", "key", key, "error", err)
	}
}

// HealthCheck returns 200 if the database is reachable, 503 otherwise.
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Ping(); err != nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// baseURL returns the scheme+host for the current request.
func (h *Handler) baseURL(reqHost string) string {
	scheme := "https"
	if strings.Contains(reqHost, ":") {
		scheme = "http"
	}
	return scheme + "://" + reqHost
}

// exampleURL builds the subdomain URL for an example site.
func (h *Handler) exampleURL(slug string) string {
	return "https://" + slug + "." + h.domain
}

// siteURL builds the public subdomain URL for a site.
func (h *Handler) siteURL(slug string) string {
	return "https://" + slug + "." + h.domain
}
