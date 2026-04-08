# Launchly Improvement Plan

## Context

Launchly is a done-for-you website service for local businesses. It currently works end-to-end (onboarding form → admin builds site → Stripe payment → live subdomain site with contact form), but has gaps in both business value and production robustness. There is no business owner portal and we're keeping it that way for now -- all improvements focus on making the done-for-you experience better and the codebase more reliable.

---

## STAGE 1: Better for Businesses

### 1.1 Welcome Email on Onboarding ✅
**Impact:** High (trust gap -- businesses submit the form and get nothing)
- **`internal/email/email.go`** -- Add `SendWelcomeEmail(to, businessName string)`. Confirm receipt, explain next steps (site built within 24h, email when ready).
- **`internal/handlers/onboarding.go`** -- Call after `CreateSite()` succeeds. Log errors, don't fail the request.

### 1.2 Spam Protection (Honeypot) ✅
**Impact:** High (contact forms have zero protection)
- **All 10 site templates** (`web/templates/sites/*.html`) -- Add hidden field: `<input type="text" name="website" style="display:none" tabindex="-1" autocomplete="off">`
- **`internal/handlers/site.go`** -- In lead submission, if `r.FormValue("website") != ""`, silently redirect as success (don't tell bot it failed)
- **`web/templates/public/onboarding.html`** + **`internal/handlers/onboarding.go`** -- Same honeypot on onboarding form

### 1.3 SEO: JSON-LD Structured Data ✅
**Impact:** High (local businesses depend on Google, zero-cost improvement)
- **`web/templates/sites/base.html`** -- Add `<script type="application/ld+json">` in `<head>` with `LocalBusiness` schema: name, description, telephone, email, address, url, image. Only include non-empty fields.
- **`internal/handlers/site.go`** -- Pass `Domain` in template data so URLs are correct

### 1.4 Click-to-Call & WhatsApp Sticky Bar ✅
**Impact:** High (most local enquiries happen by phone, especially mobile)
- **`web/templates/sites/base.html`** -- Add a `position: fixed; bottom: 0` bar visible on mobile only (`md:hidden`). Show click-to-call button if Phone is set, WhatsApp button if WhatsAppURL is set.

### 1.5 Image Upload (Cloudflare R2)
**Impact:** Highest (URL pasting is the biggest friction point for non-technical business owners)
- **New: `internal/storage/storage.go`** -- R2 client using `github.com/aws/aws-sdk-go-v2/service/s3`. Methods: `Upload(ctx, key, reader, contentType) (url, error)`, `Delete(ctx, key)`. Env vars: `R2_ACCOUNT_ID`, `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY`, `R2_BUCKET`, `R2_PUBLIC_URL`.
- **`internal/handlers/handler.go`** -- Add `storage` field to Handler struct
- **`cmd/server/main.go`** -- Init storage client, pass to Handler
- **`internal/handlers/onboarding.go`** -- Change to `multipart/form-data`. Use `r.FormFile("logo")` and `r.FormFile("gallery_files")`. Upload to R2 with keys like `sites/{slug}/logo-{uuid}.{ext}`. Max 5MB, JPEG/PNG/WebP only via `http.MaxBytesReader`.
- **`web/templates/public/onboarding.html`** -- Replace URL text inputs with `<input type="file">`. Add JS thumbnail preview.
- **`internal/handlers/admin.go`** + **`web/templates/admin/edit.html`** -- Add file upload alongside existing URL fields for admin editing
- **Dependencies:** `github.com/aws/aws-sdk-go-v2`, `github.com/google/uuid`

### 1.6 Contact Form: Service/Enquiry Type Field
**Impact:** Medium (pre-qualifies leads, makes notifications more useful)
- **`internal/models/models.go`** -- Add `Service string` to Lead
- **`internal/db/db.go`** -- Migration: `ALTER TABLE leads ADD COLUMN IF NOT EXISTS service TEXT NOT NULL DEFAULT ''`. Update CreateLead INSERT and all Lead scans.
- **All 10 site templates** -- Add `<select name="service">` dropdown populated from `{{range .Services}}`, with "General Enquiry" default
- **`internal/handlers/site.go`** -- Read `r.FormValue("service")` in lead submission
- **`internal/email/email.go`** -- Show service field in lead notification email
- **`web/templates/admin/site.html`** -- Show service in leads table

### 1.7 Subscription Lifecycle Emails ✅
**Impact:** Medium (businesses get no confirmation when they pay or cancel)
- **`internal/email/email.go`** -- Add `SendPaymentConfirmation(to, businessName, plan)` and `SendCancellationConfirmation(to, businessName)`
- **`internal/db/db.go`** -- Add `GetSiteByStripeSessionID()` and `GetSiteByStripeSubscriptionID()` methods
- **`internal/handlers/admin.go`** -- In `StripeWebhook`, after successful DB update, look up site and send appropriate email

### 1.8 Google Maps Embed ✅
**Impact:** Medium (builds trust, helps customers find the business)
- **`internal/models/models.go`** -- Add `MapEmbedURL string`
- **`internal/db/db.go`** -- Migration + CRUD updates
- **All 10 site templates** -- Add `{{if .Site.MapEmbedURL}}<iframe src="...">{{end}}` in contact section
- **`web/templates/admin/edit.html`** + **`web/templates/public/onboarding.html`** -- Add input field
- **Handlers** -- Read and persist new field

### 1.9 Lead Export (CSV) ✅
**Impact:** Low-medium (admin convenience)
- **`internal/handlers/admin.go`** -- Add `AdminExportLeads` handler. Use `encoding/csv`, set `Content-Type: text/csv` and `Content-Disposition: attachment`.
- **`web/templates/admin/site.html`** -- Add "Export CSV" link
- **Route:** `GET /admin/sites/{id}/leads.csv`

### 1.10 Weekly Lead Summary Email
**Impact:** Medium (reduces missed leads, keeps businesses engaged)
- **`internal/db/db.go`** -- Add `ListLeadsSince(siteID, since)` and `ListPaidSites()`
- **`internal/email/email.go`** -- Add `SendWeeklyLeadSummary(to, businessName, leads, siteURL)`
- **New: `internal/jobs/weekly_summary.go`** -- Iterates paid sites, fetches leads from past 7 days, sends digest
- **`cmd/server/main.go`** -- Add `--run-weekly-summary` flag. If set, run job and exit (called via Railway cron).

### 1.11 Custom Analytics (No Third-Party Service)
**Impact:** Medium (visibility into site traffic without relying on Umami or external tools)

**Why:** Replace or supplement Umami with a lightweight first-party analytics system. Stores page views and (optionally) contact form submissions in PostgreSQL. Visible in the admin panel per-site.

**New DB table:** `page_views`
- `id` SERIAL PK
- `site_id` INT FK → sites
- `path` TEXT (e.g. `/`, `/services`, etc.)
- `referrer` TEXT
- `user_agent` TEXT
- `ip_hash` TEXT (hashed, not stored raw — privacy)
- `created_at` TIMESTAMPTZ

**New files:**
- **`internal/db/db.go`** -- Add migration for `page_views` table. Add `RecordPageView(siteID, path, referrer, ipHash, userAgent)` and `GetPageViewStats(siteID, since) (total, uniqueDays int, err)`.
- **`internal/handlers/site.go`** -- In `renderSite`, fire a goroutine to record the page view (non-blocking). Hash the IP with SHA-256 before storing.

**Admin display:**
- **`internal/handlers/admin.go`** -- In `AdminSite`, query `GetPageViewStats` for the last 30 days. Pass counts to template.
- **`web/templates/admin/site.html`** -- Add a simple "Last 30 days: X views" stat card near the top of the site detail page.

**Notes:**
- No JS required — server-side only, works with ad blockers
- Exclude bots by checking User-Agent for known bot strings
- Don't record page views for draft sites
- The existing Umami integration stays as-is; this is additive

---

## STAGE 2: Production Ready

### 2.1 Stop Seeding in Production ✅
- **`cmd/server/main.go`** -- Gate `SeedExamples()` behind `if getEnv("SEED_EXAMPLES", "") == "true"`. One-line fix.

### 2.2 Graceful Shutdown + HTTP Timeouts
- **`cmd/server/main.go`** -- Replace `log.Fatal(http.ListenAndServe(...))` with `http.Server{}` struct. Set `ReadTimeout: 5s`, `WriteTimeout: 10s`, `IdleTimeout: 120s`. Use `signal.NotifyContext` for SIGTERM, call `srv.Shutdown(ctx)` with 10s deadline.

### 2.3 Structured Logging (slog)
- **`cmd/server/main.go`** -- Set up `slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))`. Update `loggingMiddleware` to use `slog.Info` with structured fields.
- **All handlers** -- Replace `log.Printf` with `slog.Info`/`slog.Error` with structured key-value fields (site ID, error, etc.)

### 2.4 Request ID Middleware
- **`cmd/server/main.go`** -- Add middleware that generates 8-char hex ID via `crypto/rand`, stores in context, sets `X-Request-ID` response header. Logging middleware includes it.

### 2.5 Panic Recovery Middleware
- **`cmd/server/main.go`** -- Add `recoveryMiddleware` with `defer func() { recover() }`. Log panic + stack trace via slog, return 500.

### 2.6 Template Caching
- **`internal/handlers/handler.go`** -- Add `templates map[string]*template.Template` to Handler. Parse all templates once in `New()` at startup.
- **All handlers** -- Replace every `template.Must(template.ParseFiles(...))` with `h.templates["name"]`. This removes the panic-on-missing-template risk entirely.

### 2.7 Security Headers Middleware
- **`cmd/server/main.go`** -- Add middleware setting: `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`, `Permissions-Policy: camera=(), microphone=(), geolocation=()`

### 2.8 Stripe Webhook: Fix Error Handling ✅
- **`internal/handlers/admin.go`** -- Return 500 (not 200) when `SetSitePaid` or `SetSiteCancelled` fails. Stripe retries on non-2xx. Critical bug fix.

### 2.9 Input Validation
- **New: `internal/validate/validate.go`** -- Helpers: `Email(s) bool`, `URL(s) bool`, `MaxLen(s, n) string`, `Phone(s) bool`
- **`internal/handlers/onboarding.go`** -- Validate lead_email format, URL fields, enforce max lengths (name: 100, tagline: 200, about: 2000)
- **`internal/handlers/site.go`** -- Validate lead email/phone
- **`internal/handlers/admin.go`** -- Validate in AdminUpdateSite

### 2.10 Rate Limiting
- **New: `internal/middleware/ratelimit.go`** -- Per-IP token bucket using `golang.org/x/time/rate`, stored in `sync.Map` with periodic cleanup. 5 req/min for form submissions.
- **`cmd/server/main.go`** -- Apply to POST routes for onboarding and contact forms. Return 429.

### 2.11 CSRF Protection
- **Dependency:** `github.com/gorilla/csrf`
- **`cmd/server/main.go`** -- Wrap mux with CSRF middleware. Exempt `/webhooks/stripe` (has its own signature verification).
- **All form templates** -- Add `{{.CSRFField}}` inside every `<form method="POST">`. Affects admin templates, onboarding, and all 10 site templates.
- **All handlers rendering forms** -- Pass CSRF token field to template data.

### 2.12 Database: Context, Pool Config, Migrations
- **`internal/db/db.go`** -- After `sql.Open`, set `SetMaxOpenConns(25)`, `SetMaxIdleConns(5)`, `SetConnMaxLifetime(5 * time.Minute)`. Change all methods to accept `ctx context.Context`, use `QueryRowContext`/`ExecContext`.
- **All handler callers** -- Pass `r.Context()` to all DB calls (mechanical change across all handlers).
- **Dependency:** `github.com/pressly/goose/v3`. Move CREATE TABLE and ALTER TABLE statements into `migrations/*.sql` files. Remove inline DDL from `db.go`.

### 2.13 Admin Auth Hardening
- **`internal/handlers/admin.go`** -- Hash admin password at startup with `golang.org/x/crypto/bcrypt`. Compare with `bcrypt.CompareHashAndPassword`. Add per-IP failed attempt tracking (5 failures → 15min lockout).

### 2.14 Tests
- **`internal/db/db.go`** -- Extract a `Querier` interface from the concrete `Store` struct. Handler depends on interface, not concrete type.
- **New test files:**
  - `internal/handlers/site_test.go` -- Test lead submission, honeypot, validation
  - `internal/handlers/onboarding_test.go` -- Test slug generation, form validation
  - `internal/validate/validate_test.go` -- Test all validation helpers
  - `internal/db/db_test.go` -- Integration tests for CRUD

---

## Implementation Order

**Stage 1** (each item is a standalone commit):
1. Welcome email (1.1) ✅
2. Spam protection (1.2) ✅
3. SEO JSON-LD (1.3) ✅
4. Click-to-call bar (1.4) ✅
5. Image upload (1.5) -- largest item
6. Contact form service field (1.6)
7. Subscription emails (1.7) ✅
8. Google Maps embed (1.8) ✅
9. Lead export CSV (1.9) ✅
10. Weekly lead summary (1.10)

**Stage 2** (order matters -- dependencies):
1. Stop seeding in prod (2.1) ✅
2. Graceful shutdown + timeouts (2.2)
3. Structured logging (2.3)
4. Request ID middleware (2.4)
5. Panic recovery (2.5)
6. Template caching (2.6)
7. Security headers (2.7)
8. Stripe webhook fix (2.8) ✅
9. Input validation (2.9)
10. Rate limiting (2.10)
11. CSRF protection (2.11)
12. DB context + pool + migrations (2.12)
13. Admin auth hardening (2.13)
14. Tests (2.14) -- depends on DB interface from 2.12

## Verification

After each item, run `go build ./...` to confirm compilation. After Stage 2 is complete, run the test suite with `go test ./...`. For Stage 1 features, manually test by running the server locally and exercising the changed flows. For Stripe changes, use test mode keys. For R2 upload, test with a dev bucket first.
