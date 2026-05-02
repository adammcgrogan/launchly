package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/adammcgrogan/launchly/internal/db"
	"github.com/adammcgrogan/launchly/internal/email"
	"github.com/adammcgrogan/launchly/internal/handlers"
	"github.com/adammcgrogan/launchly/internal/payment"
	"github.com/joho/godotenv"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
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
		slog.Error("database init failed", "error", err)
		os.Exit(1)
	}
	if err := store.Migrate(); err != nil {
		slog.Error("migrate failed", "error", err)
		os.Exit(1)
	}
	if err := store.SeedExamples(); err != nil {
		slog.Warn("seed examples failed (non-fatal)", "error", err)
	}

	mailer := email.New(resendKey, emailFrom)
	pay := payment.New(stripeSecretKey, stripeWebhookSecret, stripeStarterProduct, stripeProProduct)
	h, err := handlers.New(store, mailer, pay, domain, adminPass, umamiScriptURL)
	if err != nil {
		slog.Error("handlers init failed", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	// Static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Register app routes
	h.RegisterRoutes(mux)

	// Subdomain router: anything on *.domain hits the site handler
	// All other requests go to the main mux
	finalHandler := loggingMiddleware(subdomainRouter(domain, h, mux))

	h.StartAnalyticsCron()

	srv := &http.Server{
		Addr:         addr,
		Handler:      finalHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown on SIGTERM/SIGINT (Railway sends SIGTERM on deploy)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-quit
		slog.Info("shutting down")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("shutdown error", "error", err)
		}
	}()

	slog.Info("listening", "addr", addr, "domain", domain)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

// loggingMiddleware logs each request with method, path, status code, and duration.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start).Round(time.Millisecond),
		)
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

// subdomainRouter routes subdomain and custom-domain requests to the site handler,
// and everything else to the main mux.
func subdomainRouter(domain string, h *handlers.Handler, fallback http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := strings.ToLower(strings.Split(effectiveHost(r), ":")[0])
		isSubdomain := strings.HasSuffix(host, "."+domain)
		isLocalhost := host == "localhost" || host == "127.0.0.1"
		isMainDomain := host == domain || host == "www."+domain || isLocalhost
		isSiteDomain := isSubdomain || (!isMainDomain && host != "")

		if isSiteDomain {
			// Static assets must be served on all site domains
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
		slog.Error("required env var not set", "key", key)
		os.Exit(1)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
