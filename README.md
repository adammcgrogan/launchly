# Launchly

A done-for-you website service for local businesses. Customers fill in a form, the site gets built and reviewed, then published on a subdomain. Enquiries from the site are forwarded to the business owner by email.

Built and operated by Adam McGrogan — [launchly.ltd](https://launchly.ltd)

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

## Running locally

Copy the example env file and fill in your values:

```
cp .env.example .env
```

Run the server:

```
go run ./cmd/server/main.go
```

Build check:

```
go build ./...
```

There is no test suite currently.

## Environment variables

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string |
| `DOMAIN` | `launchly.ltd` in production |
| `ADMIN_PASSWORD` | Protects `/admin` via HTTP basic auth |
| `RESEND_API_KEY` | Resend API key |
| `EMAIL_FROM` | From address for outbound email |
| `STRIPE_SECRET_KEY` | Stripe secret key |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook signing secret |
| `STRIPE_STARTER_PRODUCT` | Stripe product ID for Starter plan |
| `STRIPE_PRO_PRODUCT` | Stripe product ID for Pro plan |
| `UMAMI_SCRIPT_URL` | Optional — Umami analytics script URL |
| `SEED_EXAMPLES` | Set to `true` to seed example sites on startup |
| `ADDR` | Server address, defaults to `:8080` |

## Site templates

Ten templates are available: `bold`, `fresh`, `warm`, `glow`, `classic`, `pulse`, `grove`, `fleet`, `haven`, `arch`. Each is a single HTML file in `web/templates/sites/`.

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
