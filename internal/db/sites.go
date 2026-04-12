package db

import (
	"database/sql"
	"time"

	"github.com/adammcgrogan/launchly/internal/models"
)

// siteColumns is the standard SELECT list for sites, used by all single-site queries.
const siteColumns = `id, slug, business_name, template, tagline, about, services,
	certifications, location, cta_text, testimonials, logo_url, gallery,
	phone, email, address, hours, map_url, map_embed_url,
	facebook_url, instagram_url, whatsapp_url, twitter_url, tiktok_url, linkedin_url, youtube_url,
	umami_website_id, lead_email, status, created_at, published_at,
	plan, payment_status, stripe_session_id, stripe_subscription_id, paid_at,
	COALESCE(custom_domain, ''), notes, analytics_frequency, analytics_last_sent`

type scanner interface {
	Scan(dest ...any) error
}

// scanSite scans a row into a Site. Returns (nil, nil) when no row exists.
func scanSite(s scanner) (*models.Site, error) {
	site := &models.Site{}
	err := s.Scan(
		&site.ID, &site.Slug, &site.BusinessName, &site.Template,
		&site.Tagline, &site.About, &site.Services,
		&site.Certifications, &site.Location, &site.CTAText,
		&site.Testimonials, &site.LogoURL, &site.Gallery,
		&site.Phone, &site.Email, &site.Address, &site.Hours, &site.MapURL, &site.MapEmbedURL,
		&site.FacebookURL, &site.InstagramURL, &site.WhatsAppURL,
		&site.TwitterURL, &site.TikTokURL, &site.LinkedInURL, &site.YouTubeURL,
		&site.UmamiWebsiteID, &site.LeadEmail, &site.Status, &site.CreatedAt, &site.PublishedAt,
		&site.Plan, &site.PaymentStatus, &site.StripeSessionID, &site.StripeSubscriptionID, &site.PaidAt,
		&site.CustomDomain, &site.Notes, &site.AnalyticsFrequency, &site.AnalyticsLastSent,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return site, err
}

func (s *Store) CreateSite(site *models.Site) error {
	return s.db.QueryRow(`
		INSERT INTO sites (slug, business_name, template, tagline, about, services,
		                   certifications, location, cta_text, testimonials, logo_url, gallery,
		                   phone, email, address, hours,
		                   map_url, map_embed_url, facebook_url, instagram_url, whatsapp_url,
		                   twitter_url, tiktok_url, linkedin_url, youtube_url,
		                   umami_website_id, lead_email, plan)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28)
		RETURNING id, created_at`,
		site.Slug, site.BusinessName, site.Template, site.Tagline, site.About,
		site.Services, site.Certifications, site.Location, site.CTAText,
		site.Testimonials, site.LogoURL, site.Gallery,
		site.Phone, site.Email, site.Address, site.Hours,
		site.MapURL, site.MapEmbedURL, site.FacebookURL, site.InstagramURL, site.WhatsAppURL,
		site.TwitterURL, site.TikTokURL, site.LinkedInURL, site.YouTubeURL,
		site.UmamiWebsiteID, site.LeadEmail, site.Plan,
	).Scan(&site.ID, &site.CreatedAt)
}

func (s *Store) GetSiteByID(id int) (*models.Site, error) {
	return scanSite(s.db.QueryRow(`SELECT `+siteColumns+` FROM sites WHERE id = $1`, id))
}

func (s *Store) GetSiteBySlug(slug string) (*models.Site, error) {
	return scanSite(s.db.QueryRow(`SELECT `+siteColumns+` FROM sites WHERE slug = $1`, slug))
}

func (s *Store) GetSiteByCustomDomain(domain string) (*models.Site, error) {
	return scanSite(s.db.QueryRow(`SELECT `+siteColumns+` FROM sites WHERE custom_domain = $1`, domain))
}

func (s *Store) GetSiteByStripeSessionID(sessionID string) (*models.Site, error) {
	return scanSite(s.db.QueryRow(`SELECT `+siteColumns+` FROM sites WHERE stripe_session_id = $1`, sessionID))
}

func (s *Store) GetSiteByStripeSubscriptionID(subID string) (*models.Site, error) {
	return scanSite(s.db.QueryRow(`SELECT `+siteColumns+` FROM sites WHERE stripe_subscription_id = $1`, subID))
}

func (s *Store) ListSites() ([]*models.Site, error) {
	rows, err := s.db.Query(`SELECT ` + siteColumns + ` FROM sites ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sites []*models.Site
	for rows.Next() {
		site, err := scanSite(rows)
		if err != nil {
			return nil, err
		}
		sites = append(sites, site)
	}
	return sites, rows.Err()
}

// LiveSiteEntry holds the minimal fields needed for sitemap generation.
type LiveSiteEntry struct {
	Slug         string
	CustomDomain string
	PublishedAt  *time.Time
}

func (s *Store) ListLiveSites() ([]LiveSiteEntry, error) {
	rows, err := s.db.Query(`
		SELECT slug, COALESCE(custom_domain, ''), published_at
		FROM sites WHERE status = 'live' ORDER BY published_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sites []LiveSiteEntry
	for rows.Next() {
		var e LiveSiteEntry
		if err := rows.Scan(&e.Slug, &e.CustomDomain, &e.PublishedAt); err != nil {
			return nil, err
		}
		sites = append(sites, e)
	}
	return sites, rows.Err()
}

func (s *Store) UpdateSite(site *models.Site) error {
	_, err := s.db.Exec(`
		UPDATE sites SET business_name=$1, tagline=$2, about=$3, services=$4,
		certifications=$5, location=$6, cta_text=$7, testimonials=$8, logo_url=$9, gallery=$10,
		phone=$11, email=$12, address=$13, hours=$14,
		map_url=$15, map_embed_url=$16, facebook_url=$17, instagram_url=$18, whatsapp_url=$19,
		twitter_url=$20, tiktok_url=$21, linkedin_url=$22, youtube_url=$23,
		umami_website_id=$24, lead_email=$25
		WHERE id=$26`,
		site.BusinessName, site.Tagline, site.About, site.Services,
		site.Certifications, site.Location, site.CTAText, site.Testimonials, site.LogoURL, site.Gallery,
		site.Phone, site.Email, site.Address, site.Hours,
		site.MapURL, site.MapEmbedURL, site.FacebookURL, site.InstagramURL, site.WhatsAppURL,
		site.TwitterURL, site.TikTokURL, site.LinkedInURL, site.YouTubeURL,
		site.UmamiWebsiteID, site.LeadEmail, site.ID,
	)
	return err
}

func (s *Store) UpdateSiteTemplate(id int, template string) error {
	_, err := s.db.Exec(`UPDATE sites SET template=$1 WHERE id=$2`, template, id)
	return err
}

func (s *Store) UpdateSiteNotes(id int, notes string) error {
	_, err := s.db.Exec(`UPDATE sites SET notes = $1 WHERE id = $2`, notes, id)
	return err
}

func (s *Store) PublishSite(id int) error {
	now := time.Now().UTC()
	_, err := s.db.Exec(`UPDATE sites SET status='live', published_at=$1 WHERE id=$2`, now, id)
	return err
}

func (s *Store) UnpublishSite(id int) error {
	_, err := s.db.Exec(`UPDATE sites SET status='draft', published_at=NULL WHERE id=$1`, id)
	return err
}

func (s *Store) DeleteSite(id int) error {
	_, err := s.db.Exec(`DELETE FROM sites WHERE id=$1`, id)
	return err
}

func (s *Store) SetSitePending(id int, plan, sessionID string) error {
	_, err := s.db.Exec(`UPDATE sites SET payment_status='pending', plan=$1, stripe_session_id=$2 WHERE id=$3`, plan, sessionID, id)
	return err
}

// SetSitePaid marks a site as paid. Returns (true, nil) if this was the first time
// (i.e. the row was updated), (false, nil) if already paid (idempotent retry).
func (s *Store) SetSitePaid(sessionID, subscriptionID string) (bool, error) {
	now := time.Now().UTC()
	res, err := s.db.Exec(`UPDATE sites SET payment_status='paid', paid_at=$1, stripe_subscription_id=$2 WHERE stripe_session_id=$3 AND payment_status != 'paid'`, now, subscriptionID, sessionID)
	if err != nil {
		return false, err
	}
	rows, _ := res.RowsAffected()
	return rows > 0, nil
}

func (s *Store) SetSiteCancelled(subscriptionID string) error {
	_, err := s.db.Exec(`UPDATE sites SET payment_status='cancelled' WHERE stripe_subscription_id=$1`, subscriptionID)
	return err
}

func (s *Store) SetCustomDomain(id int, domain string) error {
	var val *string
	if domain != "" {
		val = &domain
	}
	_, err := s.db.Exec(`UPDATE sites SET custom_domain = $1 WHERE id = $2`, val, id)
	return err
}
