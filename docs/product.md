# LocalLaunch — Product Document

> **Tagline:** Getting your business online.

---

## Overview

LocalLaunch is a done-for-you web presence service for local businesses. The business owner does nothing technical — they fill in a short onboarding form, pick a template, and get a professional landing page live within 24 hours. Leads (form submissions, calls) are forwarded directly to them. They pay a flat monthly fee and never have to touch anything.

---

## Problem

Most small local businesses have no website, or one that is outdated and ineffective. Existing solutions (Squarespace, Wix, WordPress) require time, skill, and ongoing maintenance that small business owners don't have. Agencies are expensive and slow.

---

## Solution

A streamlined, affordable, fully managed landing page service. One page, done properly, focused entirely on converting visitors into leads.

---

## Target Customers

Small local businesses across any niche, including:

- Trades (plumbers, electricians, builders, cleaners)
- Health & beauty (salons, barbers, spas, nail studios)
- Food & drink (cafés, restaurants, takeaways)
- Fitness (gyms, personal trainers, yoga studios)
- Professional services (accountants, solicitors, consultants)
- Retail & shops
- Events & entertainment

---

## Business Model

- **Monthly subscription** — flat fee per site (e.g. £49–£99/month)
- **Setup fee** (optional one-off charge to cover initial build)
- **Custom domain add-on** — business uses their own domain instead of a subdomain
- Low churn expected — once live, businesses rarely cancel

---

## How It Works (Customer Journey)

1. Business owner visits LocalLaunch and clicks **Get Started**
2. They fill in an onboarding form:
   - Business name, type, tagline
   - Services offered
   - Contact details (phone, email, address)
   - Opening hours
   - Logo and photos (optional)
   - Social media links (optional)
   - Google Maps link (optional)
3. They pick a template from 5–10 options
4. They submit — LocalLaunch receives the request
5. The site is reviewed, any final adjustments made, and published
6. The business goes live on a subdomain: `businessname.locallaunch.co`
7. Custom domain available as an add-on
8. When a visitor submits the contact form, the lead is emailed/texted directly to the business owner

---

## Templates

Each template shares the same page structure but has a distinct visual style suited to different business types. All are mobile-first and fast-loading.

| # | Name | Style | Best for |
|---|------|-------|----------|
| 1 | **Bold** | Dark, high contrast, strong typography | Trades, gyms |
| 2 | **Fresh** | Light, clean, minimal | Professional services, consultants |
| 3 | **Warm** | Earthy tones, inviting | Cafés, restaurants, food |
| 4 | **Glow** | Soft pastels, elegant | Salons, spas, beauty |
| 5 | **Energy** | Vibrant, dynamic | Fitness, personal trainers |
| 6 | **Classic** | Neutral, professional, timeless | Retail, general use |
| 7 | **Local** | Friendly, community feel | Any neighbourhood business |

More templates added over time based on demand.

---

## Landing Page Structure (All Templates)

Every page follows the same section order:

1. **Hero** — Business name, tagline, primary CTA (call or contact)
2. **About** — Short intro, who they are, what makes them different
3. **Services** — List of services offered (with optional prices)
4. **Gallery** — Photos (optional)
5. **Opening Hours** — Days and times
6. **Contact / Lead Form** — Name, phone, message — submits to business owner
7. **Location** — Address + Google Maps embed
8. **Footer** — Links, copyright, "Powered by LocalLaunch"

---

## Lead Capture

When a visitor submits the contact form:
- The lead is emailed to the business owner instantly
- Optional SMS notification
- Lead is also logged in the LocalLaunch admin dashboard
- Future: weekly lead summary email to business owner

---

## Hosting & Domains

- Every site gets a free subdomain: `businessname.locallaunch.co`
- Custom domain support as a paid add-on
- All sites served from a single Go server using subdomain routing
- SSL via Let's Encrypt (automatic)

---

## Admin Dashboard (Internal — for LocalLaunch operator)

The admin panel is used by the LocalLaunch operator (you) to manage clients. It is not accessible to business owners.

Features:
- View all client onboarding submissions
- Publish / unpublish a site
- Edit site content
- View leads per site
- Manage billing status
- Add/remove custom domains

---

## Tech Stack

| Layer | Choice | Reason |
|-------|--------|--------|
| Backend | Go | Fast, single binary, easy to deploy, great for serving many sites |
| Templating | Go `html/template` | Built-in, safe, sufficient for static-ish pages |
| Database | PostgreSQL | Reliable, simple schema |
| Email | SMTP / Resend API | Lead forwarding and notifications |
| SMS | Twilio | Optional lead SMS alerts |
| Payments | Stripe | Subscriptions and billing |
| Hosting | Single VPS (e.g. Hetzner) | Low cost, full control |
| SSL | Let's Encrypt (certmagic) | Automatic HTTPS per domain |
| File storage | S3-compatible (e.g. Backblaze B2) | Logo and photo uploads |

---

## MVP Scope

The first version ships with:

- [ ] Onboarding form (public-facing)
- [ ] Template picker (5 templates minimum)
- [ ] Site generation and subdomain routing
- [ ] Contact / lead capture form with email forwarding
- [ ] Admin dashboard (view submissions, publish sites, view leads)
- [ ] Stripe subscription billing
- [ ] SSL for all subdomains

**Out of scope for MVP:**
- Custom domain support
- SMS notifications
- Business owner login / self-editing
- Analytics / visitor stats
- Multi-language support

---

## Future Ideas

- Business owner portal (view their own leads, update opening hours)
- SEO enhancements (meta tags, schema markup, Google Business integration)
- Review widget (pull in Google reviews)
- WhatsApp contact button
- Appointment booking add-on
- Analytics dashboard (visits, leads, conversion rate)
- White-label option (agencies resell LocalLaunch)

---

## Open Questions

- What is the pricing (setup fee + monthly)?
- What markets to target first (UK, US, other)?
- What is the custom domain add-on price?
- Will there be a free trial or money-back guarantee?
