# Launchly

Check it out at [launchly.ltd](https://launchly.ltd).

A done-for-you website service for local businesses. Customers fill in a form, the site gets built and reviewed, then published on a subdomain. Enquiries from the site are forwarded to the business owner by email.

## Stack

- **Go 1.22** — single binary, standard library HTTP server
- **PostgreSQL** — via `lib/pq`, inline migrations on startup
- **html/template + Tailwind CSS** — server-rendered, no build step
- **Stripe** — subscription billing (Starter £19/mo, Pro £39/mo)
- **Resend** — transactional email
- **Cloudflare** — DNS, proxying wildcard subdomains
- **Railway** — hosting

## How it works

1. Business owner fills in `/get-started`
2. Admin reviews and builds the site via `/admin`
3. Admin publishes the site — it goes live at `slug.launchly.ltd`
4. Admin sends a Stripe payment link
5. Visitor submits the contact form — lead is saved and emailed to the business

## Site templates

13 templates are available: `bold`, `fresh`, `warm`, `glow`, `classic`, `pulse`, `grove`, `fleet`, `haven`, `arch`, `dine`, `heal`, `craft`. Each is a single HTML file in `web/templates/sites/`.

## Project structure

```
cmd/server/main.go        — entry point, routing, middleware
internal/
  handlers/               — HTTP handlers (admin, site, onboarding, home)
  db/db.go                — all database queries and migrations
  email/email.go          — outbound email via Resend
  payment/payment.go      — Stripe checkout and webhook handling
  models/                 — shared data structs
web/
  templates/              — HTML templates (public, admin, sites)
  static/                 — CSS and images
```
