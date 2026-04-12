package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Store struct {
	db *sql.DB
}

func New(dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	// Conservative pool limits — Railway free/hobby Postgres caps connections.
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(5 * time.Minute)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("cannot connect to database: %w", err)
	}
	return &Store{db: db}, nil
}

// Ping checks that the database is reachable. Used by the health check endpoint.
func (s *Store) Ping() error {
	return s.db.Ping()
}

func (s *Store) Migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS sites (
			id               SERIAL PRIMARY KEY,
			slug             TEXT NOT NULL UNIQUE,
			business_name    TEXT NOT NULL,
			template         TEXT NOT NULL DEFAULT 'bold',
			tagline          TEXT NOT NULL DEFAULT '',
			about            TEXT NOT NULL DEFAULT '',
			services         TEXT NOT NULL DEFAULT '',
			certifications   TEXT NOT NULL DEFAULT '',
			location         TEXT NOT NULL DEFAULT '',
			cta_text         TEXT NOT NULL DEFAULT '',
			testimonials     TEXT NOT NULL DEFAULT '',
			logo_url         TEXT NOT NULL DEFAULT '',
			gallery          TEXT NOT NULL DEFAULT '',
			phone            TEXT NOT NULL DEFAULT '',
			email            TEXT NOT NULL DEFAULT '',
			address          TEXT NOT NULL DEFAULT '',
			hours            TEXT NOT NULL DEFAULT '',
			map_url          TEXT NOT NULL DEFAULT '',
			map_embed_url    TEXT NOT NULL DEFAULT '',
			facebook_url     TEXT NOT NULL DEFAULT '',
			instagram_url    TEXT NOT NULL DEFAULT '',
			whatsapp_url     TEXT NOT NULL DEFAULT '',
			twitter_url      TEXT NOT NULL DEFAULT '',
			tiktok_url       TEXT NOT NULL DEFAULT '',
			linkedin_url     TEXT NOT NULL DEFAULT '',
			youtube_url      TEXT NOT NULL DEFAULT '',
			umami_website_id TEXT NOT NULL DEFAULT '',
			lead_email       TEXT NOT NULL,
			status           TEXT NOT NULL DEFAULT 'draft',
			created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			published_at     TIMESTAMPTZ,
			plan                    TEXT NOT NULL DEFAULT '',
			payment_status          TEXT NOT NULL DEFAULT 'unpaid',
			stripe_session_id       TEXT NOT NULL DEFAULT '',
			stripe_subscription_id  TEXT NOT NULL DEFAULT '',
			paid_at                 TIMESTAMPTZ
		);

		CREATE TABLE IF NOT EXISTS leads (
			id         SERIAL PRIMARY KEY,
			site_id    INT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
			name       TEXT NOT NULL,
			email      TEXT NOT NULL DEFAULT '',
			phone      TEXT NOT NULL DEFAULT '',
			message    TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return err
	}
	// Add new columns for existing installs (idempotent)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS certifications TEXT NOT NULL DEFAULT ''`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS location TEXT NOT NULL DEFAULT ''`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS cta_text TEXT NOT NULL DEFAULT ''`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS testimonials TEXT NOT NULL DEFAULT ''`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS logo_url TEXT NOT NULL DEFAULT ''`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS gallery TEXT NOT NULL DEFAULT ''`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS whatsapp_url TEXT NOT NULL DEFAULT ''`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS twitter_url TEXT NOT NULL DEFAULT ''`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS tiktok_url TEXT NOT NULL DEFAULT ''`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS linkedin_url TEXT NOT NULL DEFAULT ''`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS youtube_url TEXT NOT NULL DEFAULT ''`)
	s.db.Exec(`ALTER TABLE leads ADD COLUMN IF NOT EXISTS email TEXT NOT NULL DEFAULT ''`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS umami_website_id TEXT NOT NULL DEFAULT ''`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS plan TEXT NOT NULL DEFAULT ''`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS payment_status TEXT NOT NULL DEFAULT 'unpaid'`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS stripe_session_id TEXT NOT NULL DEFAULT ''`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS stripe_subscription_id TEXT NOT NULL DEFAULT ''`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS paid_at TIMESTAMPTZ`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS map_embed_url TEXT NOT NULL DEFAULT ''`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS custom_domain TEXT`)
	s.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_sites_custom_domain ON sites (custom_domain) WHERE custom_domain IS NOT NULL AND custom_domain != ''`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT ''`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS analytics_frequency TEXT NOT NULL DEFAULT 'off'`)
	s.db.Exec(`ALTER TABLE sites ADD COLUMN IF NOT EXISTS analytics_last_sent TIMESTAMPTZ`)
	s.db.Exec(`CREATE TABLE IF NOT EXISTS page_views (
		id         SERIAL PRIMARY KEY,
		site_id    INT NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
		viewed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		path       TEXT NOT NULL DEFAULT '/',
		referrer   TEXT NOT NULL DEFAULT '',
		ip         TEXT NOT NULL DEFAULT '',
		user_agent TEXT NOT NULL DEFAULT '',
		country    TEXT NOT NULL DEFAULT ''
	)`)
	s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_page_views_site_time ON page_views (site_id, viewed_at)`)
	return nil
}
