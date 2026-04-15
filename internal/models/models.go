package models

import "time"

type SiteStatus string

const (
	StatusDraft SiteStatus = "draft"
	StatusLive  SiteStatus = "live"
)

// Site represents a business's landing page.
type Site struct {
	ID               int        `db:"id"`
	Slug             string     `db:"slug"`             // used for subdomain: slug.launchly.ltd
	BusinessName     string     `db:"business_name"`
	Template         string     `db:"template"`
	Tagline          string     `db:"tagline"`
	About            string     `db:"about"`
	Services         string     `db:"services"`         // newline-separated list
	Certifications   string     `db:"certifications"`   // newline-separated trust badges e.g. "Gas Safe Registered"
	Location         string     `db:"location"`         // short location for hero badge e.g. "Belfast, NI"
	CTAText          string     `db:"cta_text"`         // primary call-to-action button text e.g. "Get a Quote"
	Testimonials     string     `db:"testimonials"`     // newline-separated "Name|Role|Quote"
	LogoURL          string     `db:"logo_url"`         // URL to business logo image
	Gallery          string     `db:"gallery"`          // newline-separated image URLs
	Phone            string     `db:"phone"`
	Email            string     `db:"email"`
	Address          string     `db:"address"`
	Hours            string     `db:"hours"`            // newline-separated e.g. "Mon-Fri: 9am-5pm"
	MapURL           string     `db:"map_url"`
	MapEmbedURL      string     `db:"map_embed_url"`
	FacebookURL      string     `db:"facebook_url"`
	InstagramURL     string     `db:"instagram_url"`
	WhatsAppURL      string     `db:"whatsapp_url"`
	TwitterURL       string     `db:"twitter_url"`
	TikTokURL        string     `db:"tiktok_url"`
	LinkedInURL      string     `db:"linkedin_url"`
	YouTubeURL       string     `db:"youtube_url"`
	UmamiWebsiteID   string     `db:"umami_website_id"`   // Umami analytics website ID
	LeadEmail        string     `db:"lead_email"`         // where leads are forwarded
	Status           SiteStatus `db:"status"`
	CreatedAt        time.Time  `db:"created_at"`
	PublishedAt      *time.Time `db:"published_at"`
	Plan                   string     `db:"plan"`                    // starter, pro
	PaymentStatus          string     `db:"payment_status"`          // unpaid, pending, paid, cancelled
	StripeSessionID        string     `db:"stripe_session_id"`
	StripeSubscriptionID   string     `db:"stripe_subscription_id"`
	PaidAt                 *time.Time `db:"paid_at"`
	CustomDomain           string     `db:"custom_domain"`           // e.g. "joesbarbershop.com" (Pro only)
	Notes                  string     `db:"notes"`                   // internal admin notes, never shown to customer
	AnalyticsFrequency     string     `db:"analytics_frequency"`     // "off", "weekly", "monthly"
	AnalyticsLastSent      *time.Time `db:"analytics_last_sent"`
	Palette                string     `db:"palette"`                 // selected colour palette ID, empty = template default
	HeadingFont            string     `db:"heading_font"`            // "sans", "serif", "mono" — empty = sans
}

// ReferrerCount is a referrer hostname with its visit count.
type ReferrerCount struct {
	Referrer string
	Count    int
}

// DayCount is a single day's view count.
type DayCount struct {
	Day   time.Time
	Count int
}

// SiteStats holds aggregated analytics for a site over a period.
type SiteStats struct {
	TotalViews     int
	UniqueVisitors int
	TopReferrers   []ReferrerCount
	ViewsByDay     []DayCount
	PeriodDays     int
}

// Lead represents a contact form submission from a site visitor.
type Lead struct {
	ID        int       `db:"id"`
	SiteID    int       `db:"site_id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Phone     string    `db:"phone"`
	Message   string    `db:"message"`
	CreatedAt time.Time `db:"created_at"`
}
