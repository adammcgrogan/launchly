# Launchly

Check it out at [launchly.ltd](https://launchly.ltd).

A done-for-you website service for local businesses. Customers fill in a form, the site gets built and published on a subdomain within 24 hours. Enquiries from the site are forwarded to the business owner by email.

## Stack

- **Go 1.22** — single binary, standard library HTTP server
- **PostgreSQL** — via `lib/pq`, inline migrations on startup
- **html/template + Tailwind CSS** — server-rendered, no build step
- **Stripe** — subscription billing (Starter £14/mo, Pro £34/mo)
- **Resend** — transactional email
- **Cloudflare** — DNS, proxying wildcard subdomains
- **Railway** — hosting

## How it works

1. Business owner fills in `/get-started` — 14-day free trial, no card required
2. Adam builds and publishes the site via `/admin`
3. Site goes live at `slug.launchly.ltd` (or a custom domain on Pro)
4. Admin sends a Stripe payment link when the trial ends
5. Visitor submits the contact form — lead is saved and emailed to the business

## Site templates

21 templates across general and industry-specific categories:

**General:** `canvas`, `slate`, `hearth`, `linen`, `onyx`, `ink`, `copper`, `builder`

**Industry:** `salon`, `gym`, `landscaping`, `garage`, `bnb`, `restaurant`, `clinic`, `maker`, `retail`, `wedding`, `barber`, `takeaway`

Each is a single HTML file in `web/templates/sites/`.

## Project structure

```
cmd/server/main.go        — entry point, routing, middleware
internal/
  handlers/               — HTTP handlers (admin, site, onboarding, home, catalog, seo)
  db/
    db.go                 — connection, migrations, seeding
    sites.go              — site queries
    leads.go              — lead queries
    analytics.go          — page view and analytics queries
  email/email.go          — outbound email via Resend
  payment/payment.go      — Stripe checkout and webhook handling
  models/                 — shared data structs
web/
  templates/              — HTML templates (public, admin, sites)
  static/                 — CSS and images
```
