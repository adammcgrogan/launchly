# Launchly — Product Document

> **Tagline:** Done-for-you websites for local businesses.

---

## Overview

Launchly is a done-for-you web presence service for local businesses. The business owner does nothing technical — they fill in a short onboarding form, pick a template, and get a professional landing page live within 24 hours. Leads from the contact form are forwarded directly to them by email. They pay a flat monthly subscription and never have to touch anything.

Operated by Adam McGrogan — [launchly.ltd](https://launchly.ltd)

---

## Problem

Most small local businesses have no website, or one that is outdated and ineffective. Existing solutions (Squarespace, Wix, WordPress) require time, skill, and ongoing maintenance. Agencies are expensive and slow.

---

## Solution

A streamlined, affordable, fully managed landing page service. One page, done properly, focused entirely on converting visitors into leads.

---

## Target Customers

Small local businesses across any niche, including:

- Trades (plumbers, electricians, builders, cleaners)
- Health & beauty (salons, barbers, spas, nail studios)
- Food & drink (cafés, restaurants, takeaways, bakeries)
- Fitness (gyms, personal trainers, yoga studios)
- Professional services (accountants, solicitors, consultants)
- Retail & shops
- Hospitality (B&Bs, guesthouses)

---

## Pricing

| Plan | Price | Build time |
|------|-------|------------|
| Starter | £19/mo | Within 24 hours |
| Pro | £39/mo | Within 12 hours |

- 14-day free trial, no card required upfront
- No contracts, cancel anytime
- Custom domain included on Pro (set up manually via Cloudflare)

---

## How It Works (Customer Journey)

1. Business owner visits Launchly and clicks **Get Started**
2. They fill in a 4-step onboarding form:
   - Business name, tagline, about, services, trust badges
   - Contact details (phone, email, address, hours, social links)
   - Logo and gallery photos (URL paste; file upload planned)
   - Template choice and plan selection
3. They submit — a welcome email confirms receipt
4. Adam reviews the submission, makes any adjustments, and publishes
5. The business goes live at `slug.launchly.ltd`
6. Adam sends a Stripe payment link to the business
7. Customer pays — subscription begins, payment confirmation email sent
8. When a visitor submits the contact form, the lead is emailed to the business instantly

---

## Templates

13 templates available, each mobile-first with a distinct visual style:

| Name | Style | Best for |
|------|-------|----------|
| Bold | Dark, high contrast, strong typography | Trades, gyms |
| Fresh | Light, clean, minimal | Professional services |
| Warm | Earthy tones, inviting | Cafés, bakeries, food |
| Glow | Soft pastels, elegant | Salons, beauty, spas |
| Classic | Neutral, professional, timeless | General use |
| Pulse | High energy, neon accents | Fitness, gyms |
| Grove | Natural greens, earthy | Landscaping, outdoors |
| Fleet | Dark industrial, orange accents | Auto, mechanics, trades |
| Haven | Warm neutral, calm | B&Bs, hospitality |
| Arch | Editorial, minimal serif | Interiors, design, studios |
| Dine | Dark, moody, menu-style layout | Restaurants, pubs, takeaways |
| Heal | Clean, clinical, trust-focused | Dentists, physios, clinics |
| Craft | Earthy, artisan, gallery-first | Makers, bakers, studios |

---

## Landing Page Structure

Every page follows the same section order:

1. **Hero** — Business name, tagline, location badge, primary CTA
2. **Services** — Grid of services offered
3. **About** — Short intro paragraph
4. **Contact form** — Name, phone, email, message → lead forwarded by email
5. **Opening hours + address** — With optional Google Maps embed
6. **Social links**
7. **Testimonials** — Up to 3 customer quotes (optional)
8. **Gallery** — Photo grid (optional)
9. **Footer** — Address, "Powered by Launchly"

Mobile sticky bar: click-to-call and WhatsApp buttons fixed at the bottom on mobile if phone/WhatsApp is set.

---

## Lead Capture

When a visitor submits the contact form:

- Lead is saved to the database
- Email notification sent to the business owner instantly (with reply-to set to the visitor's email)
- All leads visible in the admin panel with CSV export
- Weekly lead summary email (planned)

---

## SEO

Each site includes:

- Full meta tags (title, description, Open Graph, Twitter card)
- Canonical URL
- JSON-LD `LocalBusiness` structured data (name, phone, email, address, logo, URL)

---

## Hosting & Infrastructure

- All sites hosted on Railway
- Subdomain routing via Cloudflare Worker — `slug.launchly.ltd`
- SSL via Cloudflare (automatic)
- Single Go binary serves all sites

---

## Admin Dashboard

Internal panel for the Launchly operator. Not accessible to business owners.

- View all onboarding submissions
- Edit site content
- Switch template
- Publish / unpublish
- View and export leads per site
- Send Stripe payment link
- Cancel subscription

---

## Tech Stack

| Layer | Choice |
|-------|--------|
| Backend | Go 1.22 |
| Templating | `html/template` + Tailwind CSS (CDN) |
| Database | PostgreSQL via `lib/pq` |
| Email | Resend API |
| Payments | Stripe (subscriptions) |
| Hosting | Railway |
| DNS / Proxy | Cloudflare |
| Analytics | Umami (optional per site) |

---

## Out of Scope (Current)

- Business owner login or self-editing
- SMS notifications
- Image file uploads (URL paste only — R2 upload planned)
- Custom domain automation (manual via Cloudflare for now)
- Multi-page sites
- Appointment booking

---

## Planned / Future

- Image upload via Cloudflare R2 (replaces URL pasting)
- First-party analytics — page views stored in PostgreSQL, visible in admin
- Weekly lead summary email to business owners
- Contact form enquiry type field (pre-qualifies leads)
- Business owner portal (read-only: view leads, site status)
- Google Reviews widget
- White-label option
