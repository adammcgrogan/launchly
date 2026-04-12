package db

import (
	"database/sql"
	"fmt"

	"github.com/adammcgrogan/launchly/internal/models"
)

func (s *Store) CreateProspect(p *models.Prospect) error {
	return s.db.QueryRow(`
		INSERT INTO prospects (business_name, trade, location, phone, email, website, source, status, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at`,
		p.BusinessName, p.Trade, p.Location, p.Phone, p.Email, p.Website, p.Source, p.Status, p.Notes,
	).Scan(&p.ID, &p.CreatedAt)
}

func (s *Store) ListProspects(status string) ([]*models.Prospect, error) {
	var rows *sql.Rows
	var err error
	if status != "" {
		rows, err = s.db.Query(`
			SELECT id, business_name, trade, location, phone, email, website, source, status, notes, created_at, contacted_at
			FROM prospects WHERE status = $1 ORDER BY created_at DESC`, status)
	} else {
		rows, err = s.db.Query(`
			SELECT id, business_name, trade, location, phone, email, website, source, status, notes, created_at, contacted_at
			FROM prospects ORDER BY created_at DESC`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.Prospect
	for rows.Next() {
		p := &models.Prospect{}
		if err := rows.Scan(&p.ID, &p.BusinessName, &p.Trade, &p.Location, &p.Phone, &p.Email,
			&p.Website, &p.Source, &p.Status, &p.Notes, &p.CreatedAt, &p.ContactedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) GetProspectByID(id int) (*models.Prospect, error) {
	p := &models.Prospect{}
	err := s.db.QueryRow(`
		SELECT id, business_name, trade, location, phone, email, website, source, status, notes, created_at, contacted_at
		FROM prospects WHERE id = $1`, id,
	).Scan(&p.ID, &p.BusinessName, &p.Trade, &p.Location, &p.Phone, &p.Email,
		&p.Website, &p.Source, &p.Status, &p.Notes, &p.CreatedAt, &p.ContactedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func (s *Store) UpdateProspect(p *models.Prospect) error {
	_, err := s.db.Exec(`
		UPDATE prospects SET
			business_name = $1, trade = $2, location = $3, phone = $4, email = $5,
			website = $6, source = $7, status = $8, notes = $9, contacted_at = $10
		WHERE id = $11`,
		p.BusinessName, p.Trade, p.Location, p.Phone, p.Email,
		p.Website, p.Source, p.Status, p.Notes, p.ContactedAt, p.ID,
	)
	return err
}

func (s *Store) DeleteProspect(id int) error {
	res, err := s.db.Exec(`DELETE FROM prospects WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("prospect %d not found", id)
	}
	return nil
}
