# Launchly — Project Guide

Launchly is a done-for-you website service for local businesses. Customers fill in a form, Adam builds their site, and they start receiving enquiries. Subscriptions are managed via Stripe.

---

## Stack

- **Language:** Go 1.22
- **Frontend:** `html/template` + Tailwind CSS v3 (CDN, no build step)
- **Database:** PostgreSQL via `lib/pq`
- **Hosting:** Railway
- **DNS / Proxy:** Cloudflare
- **Email:** Resend (transactional, from `info@launchly.ltd`)
- **Payments:** Stripe (subscriptions only — always use `mode: subscription`, never `mode: payment`)

---

## Project Structure

```
cmd/server/main.go          — entry point, env vars, routing setup
internal/
  handlers/
    admin.go                — admin dashboard + RegisterRoutes()
    handler.go              — Handler struct, helpers (siteURL, exampleURL)
    home.go                 — public homepage
    onboarding.go           — /get-started form
    site.go                 — subdomain site serving + lead submission
    seo.go                  — robots.txt, sitemap.xml, privacy, terms
  db/db.go                  — all database queries
  email/email.go            — all outbound emails (Resend)
  payment/payment.go        — Stripe checkout, cancel, webhook parsing
  models/                   — shared data models
web/
  templates/
    public/                 — homepage, onboarding, legal pages (use home_base.html)
    admin/                  — admin panel templates
    sites/                  — business site templates (bold, fresh, warm, etc.)
  static/                   — CSS, images
```

---

## Environment Variables

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string |
| `DOMAIN` | `launchly.ltd` (no trailing slash, no https://) |
| `ADMIN_PASSWORD` | Protects /admin via basic auth |
| `RESEND_API_KEY` | Resend API key for sending emails |
| `EMAIL_FROM` | `Launchly <info@launchly.ltd>` |
| `STRIPE_SECRET_KEY` | `sk_live_...` in production, `sk_test_...` for testing |
| `STRIPE_WEBHOOK_SECRET` | `whsec_...` from Stripe webhook endpoint |
| `STRIPE_STARTER_PRODUCT` | Stripe product ID for Starter plan (£19/mo) |
| `STRIPE_PRO_PRODUCT` | Stripe product ID for Pro plan (£39/mo) |
| `UMAMI_SCRIPT_URL` | Optional — Umami analytics script URL |
| `ADDR` | Default `:8080` |

---

## Subdomain Routing

Business sites are served via subdomain: `slug.launchly.ltd`

- Cloudflare Worker intercepts `*.launchly.ltd/*` and sets `X-Real-Host` header
- Railway proxies to the Go app (overwrites `X-Forwarded-Host`, so we use `X-Real-Host`)
- `effectiveHost()` in `main.go` reads `X-Real-Host` first, falls back to `r.Host`
- `extractSlug()` in `site.go` parses the slug from the host
- Path-based `/sites/{slug}` routes are kept as a fallback for local development

---

## Templates

All public pages use `home_base.html` as the base template (Tailwind). Do not use `base.html` (old, CSS-only).

Business site templates are in `web/templates/sites/`. Available templates:
`bold`, `fresh`, `warm`, `glow`, `classic`, `pulse`, `grove`, `fleet`, `haven`, `arch`

Each template defines two blocks: `{{define "theme"}}` and `{{define "site-content"}}`.

---

## Email

All emails go through `internal/email/email.go`. Each email type has its own method. All are wrapped with `wrap()` which applies the branded HTML shell (indigo header, white body, grey footer).

Transactional emails send from `info@launchly.ltd`. Customer-facing contact uses `hello@launchly.ltd` (Cloudflare Email Routing → `adammcgrogan2005@gmail.com`).

---

## Stripe

- All plans are **recurring subscriptions** — always use `mode: subscription` in checkout sessions
- Webhook events handled: `checkout.session.completed`, `customer.subscription.deleted`
- Webhook endpoint: `https://launchly.ltd/webhooks/stripe`
- If a subscription is missing from Stripe (cancelled externally), treat it as already cancelled — do not error

---

## Payments Flow

1. Admin sends payment link via `/admin/sites/{id}/send-payment`
2. Creates Stripe Checkout session → sends URL to customer's email
3. Customer pays → Stripe fires `checkout.session.completed` webhook
4. App marks site as paid and stores subscription ID
5. Cancellation via admin or Stripe dashboard → `customer.subscription.deleted` webhook updates DB

---

## Key Business Details

- **Plans:** Starter £19/mo, Pro £39/mo
- **Free trial:** 14 days, no card required upfront
- **Build guarantee:** within 24 hours of approval (Pro: 12 hours)
- **Domain:** launchly.ltd — owned by Adam McGrogan, Belfast
- **Founder contact:** hello@launchly.ltd / adammcgrogan2005@gmail.com
- **No contracts, cancel anytime**

---

## Local Development

```bash
go build ./...        # build check
go run cmd/server/main.go  # run locally (requires .env file)
```

There is no test suite currently. Use `go build ./...` to verify changes compile.
