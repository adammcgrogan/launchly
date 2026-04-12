package db

import "github.com/adammcgrogan/launchly/internal/models"

func (s *Store) CreateLead(lead *models.Lead) error {
	return s.db.QueryRow(`
		INSERT INTO leads (site_id, name, email, phone, message)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`,
		lead.SiteID, lead.Name, lead.Email, lead.Phone, lead.Message,
	).Scan(&lead.ID, &lead.CreatedAt)
}

func (s *Store) ListLeadsBySite(siteID int) ([]*models.Lead, error) {
	rows, err := s.db.Query(`
		SELECT id, site_id, name, email, phone, message, created_at
		FROM leads WHERE site_id = $1 ORDER BY created_at DESC`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var leads []*models.Lead
	for rows.Next() {
		l := &models.Lead{}
		if err := rows.Scan(&l.ID, &l.SiteID, &l.Name, &l.Email, &l.Phone, &l.Message, &l.CreatedAt); err != nil {
			return nil, err
		}
		leads = append(leads, l)
	}
	return leads, rows.Err()
}

func (s *Store) ListAllLeads() ([]*models.Lead, error) {
	rows, err := s.db.Query(`
		SELECT id, site_id, name, email, phone, message, created_at
		FROM leads ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var leads []*models.Lead
	for rows.Next() {
		l := &models.Lead{}
		if err := rows.Scan(&l.ID, &l.SiteID, &l.Name, &l.Email, &l.Phone, &l.Message, &l.CreatedAt); err != nil {
			return nil, err
		}
		leads = append(leads, l)
	}
	return leads, rows.Err()
}
