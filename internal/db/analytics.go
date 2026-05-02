package db

import (
	"fmt"
	"time"

	"github.com/adammcgrogan/launchly/internal/models"
)

func (s *Store) RecordPageView(siteID int, path, referrer, ip, userAgent, country string) error {
	_, err := s.db.Exec(`
		INSERT INTO page_views (site_id, path, referrer, ip, user_agent, country)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		siteID, path, referrer, ip, userAgent, country)
	return err
}

func (s *Store) GetSiteStats(siteID int, since time.Time) (*models.SiteStats, error) {
	stats := &models.SiteStats{
		PeriodDays: int(time.Since(since).Hours()/24) + 1,
	}

	s.db.QueryRow(`SELECT COUNT(*) FROM page_views WHERE site_id=$1 AND viewed_at > $2`, siteID, since).Scan(&stats.TotalViews)
	s.db.QueryRow(`SELECT COUNT(DISTINCT ip) FROM page_views WHERE site_id=$1 AND viewed_at > $2`, siteID, since).Scan(&stats.UniqueVisitors)

	rows, err := s.db.Query(`
		SELECT referrer, COUNT(*) AS cnt
		FROM page_views
		WHERE site_id=$1 AND viewed_at > $2 AND referrer != ''
		GROUP BY referrer ORDER BY cnt DESC LIMIT 5`, siteID, since)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var rc models.ReferrerCount
			rows.Scan(&rc.Referrer, &rc.Count)
			stats.TopReferrers = append(stats.TopReferrers, rc)
		}
	}

	rows2, err := s.db.Query(`
		SELECT date_trunc('day', viewed_at AT TIME ZONE 'UTC') AS day, COUNT(*) AS cnt
		FROM page_views
		WHERE site_id=$1 AND viewed_at > $2
		GROUP BY day ORDER BY day`, siteID, since)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var dc models.DayCount
			rows2.Scan(&dc.Day, &dc.Count)
			stats.ViewsByDay = append(stats.ViewsByDay, dc)
		}
	}

	return stats, nil
}

func (s *Store) GetSitesDueForAnalytics() ([]*models.Site, error) {
	rows, err := s.db.Query(`
		SELECT id, slug, business_name, lead_email, analytics_frequency
		FROM sites
		WHERE analytics_frequency != 'off'
		  AND lead_email != ''
		  AND (
		    analytics_last_sent IS NULL
		    OR (analytics_frequency = 'weekly'  AND analytics_last_sent < NOW() - INTERVAL '7 days')
		    OR (analytics_frequency = 'monthly' AND analytics_last_sent < NOW() - INTERVAL '30 days')
		  )`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sites []*models.Site
	for rows.Next() {
		s := &models.Site{}
		if err := rows.Scan(&s.ID, &s.Slug, &s.BusinessName, &s.LeadEmail, &s.AnalyticsFrequency); err != nil {
			return nil, err
		}
		sites = append(sites, s)
	}
	return sites, rows.Err()
}

func (s *Store) UpdateAnalyticsFrequency(id int, freq string) error {
	_, err := s.db.Exec(`UPDATE sites SET analytics_frequency = $1 WHERE id = $2`, freq, id)
	return err
}

func (s *Store) UpdateAnalyticsLastSent(id int) error {
	_, err := s.db.Exec(`UPDATE sites SET analytics_last_sent = NOW() WHERE id = $1`, id)
	return err
}

// GetSitesDueForTrialReminder returns unpaid sites whose trial is ending soon and
// haven't yet received a reminder of the given type ("first" = day 11, "final" = day 13).
func (s *Store) GetSitesDueForTrialReminder(reminderType string) ([]*models.Site, error) {
	var query string
	switch reminderType {
	case "first":
		// Send on day 11 (≤3 days left), if first reminder not yet sent
		query = `SELECT ` + siteColumns + ` FROM sites
			WHERE trial_ends_at IS NOT NULL
			  AND payment_status NOT IN ('paid', 'cancelled')
			  AND lead_email != ''
			  AND trial_ends_at <= NOW() + INTERVAL '3 days'
			  AND trial_ends_at > NOW()
			  AND trial_reminder_sent IS NULL`
	case "final":
		// Send on day 13 (≤1 day left), if final reminder not yet sent
		query = `SELECT ` + siteColumns + ` FROM sites
			WHERE trial_ends_at IS NOT NULL
			  AND payment_status NOT IN ('paid', 'cancelled')
			  AND lead_email != ''
			  AND trial_ends_at <= NOW() + INTERVAL '1 day'
			  AND trial_ends_at > NOW()
			  AND trial_final_reminder_sent IS NULL`
	default:
		return nil, fmt.Errorf("unknown reminder type: %s", reminderType)
	}
	rows, err := s.db.Query(query)
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

// MarkTrialReminderSent records that a trial reminder was sent for a site.
func (s *Store) MarkTrialReminderSent(id int, reminderType string) error {
	var col string
	switch reminderType {
	case "first":
		col = "trial_reminder_sent"
	case "final":
		col = "trial_final_reminder_sent"
	default:
		return fmt.Errorf("unknown reminder type: %s", reminderType)
	}
	_, err := s.db.Exec(`UPDATE sites SET `+col+` = NOW() WHERE id = $1`, id)
	return err
}

// MarkStripeEventProcessed records a Stripe event ID. Returns true if newly inserted
// (first delivery), false if the event was already processed (retry/duplicate).
func (s *Store) MarkStripeEventProcessed(eventID string) (bool, error) {
	res, err := s.db.Exec(
		`INSERT INTO stripe_events (event_id) VALUES ($1) ON CONFLICT (event_id) DO NOTHING`,
		eventID,
	)
	if err != nil {
		return false, err
	}
	rows, _ := res.RowsAffected()
	return rows > 0, nil
}
