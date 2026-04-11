package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/adammcgrogan/locallaunch/internal/db"
	"github.com/adammcgrogan/locallaunch/internal/email"
	"github.com/adammcgrogan/locallaunch/internal/handlers"
	"github.com/adammcgrogan/locallaunch/internal/payment"
	"github.com/joho/godotenv"
)

func main() {
	log.SetOutput(os.Stdout)
	_ = godotenv.Load()

	dsn := mustEnv("DATABASE_URL")
	domain := getEnv("DOMAIN", "launchly.ltd")
	adminPass := mustEnv("ADMIN_PASSWORD")
	resendKey := getEnv("RESEND_API_KEY", "")
	emailFrom := getEnv("EMAIL_FROM", "noreply@launchly.ltd")
	umamiScriptURL := getEnv("UMAMI_SCRIPT_URL", "")
	stripeSecretKey := getEnv("STRIPE_SECRET_KEY", "")
	stripeWebhookSecret := getEnv("STRIPE_WEBHOOK_SECRET", "")
	stripeStarterProduct := getEnv("STRIPE_STARTER_PRODUCT", "")
	stripeProProduct := getEnv("STRIPE_PRO_PRODUCT", "")
	addr := getEnv("ADDR", ":8080")

	store, err := db.New(dsn)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	if err := store.Migrate(); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := store.SeedExamples(); err != nil {
		log.Fatalf("seed examples: %v", err)
	}

	mailer := email.New(resendKey, emailFrom)
	pay := payment.New(stripeSecretKey, stripeWebhookSecret, stripeStarterProduct, stripeProProduct)
	h := handlers.New(store, mailer, pay, domain, adminPass, umamiScriptURL)

	mux := http.NewServeMux()

	// Static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Register app routes
	h.RegisterRoutes(mux)

	// Subdomain router: anything on *.domain hits the site handler
	// All other requests go to the main mux
	finalHandler := loggingMiddleware(subdomainRouter(domain, h, mux))

	log.Printf("Launchly listening on %s (domain: %s)", addr, domain)
	log.Fatal(http.ListenAndServe(addr, finalHandler))
}

// loggingMiddleware logs each request with method, path, status code, and duration.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rec.status, time.Since(start).Round(time.Millisecond))
	})
}

// statusRecorder wraps ResponseWriter to capture the status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// effectiveHost returns X-Forwarded-Host if set (e.g. from a Cloudflare Worker
// proxying wildcard subdomains), falling back to the raw Host header.
func effectiveHost(r *http.Request) string {
	if fh := r.Header.Get("X-Real-Host"); fh != "" {
		return fh
	}
	return r.Host
}

// subdomainRouter routes subdomain requests to the site handler,
// and everything else to the main mux.
func subdomainRouter(domain string, h *handlers.Handler, fallback http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := strings.ToLower(strings.Split(effectiveHost(r), ":")[0])
		if strings.HasSuffix(host, "."+domain) {
			// Static assets must be served on subdomains too
			if strings.HasPrefix(r.URL.Path, "/static/") {
				fallback.ServeHTTP(w, r)
				return
			}
			// Contact form on business sites
			if r.Method == http.MethodPost && r.URL.Path == "/contact" {
				h.SubmitLead(w, r)
				return
			}
			h.ServeSite(w, r)
			return
		}
		fallback.ServeHTTP(w, r)
	})
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
