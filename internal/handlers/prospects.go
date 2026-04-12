package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/adammcgrogan/launchly/internal/models"
)

var prospectStatuses = []string{"new", "contacted", "interested", "won", "lost"}

// AdminAPICreateProspect accepts JSON from external tools (e.g. the /outreach skill).
func (h *Handler) AdminAPICreateProspect(w http.ResponseWriter, r *http.Request) {
	var body struct {
		BusinessName string `json:"business_name"`
		Trade        string `json:"trade"`
		Location     string `json:"location"`
		Phone        string `json:"phone"`
		Email        string `json:"email"`
		Website      string `json:"website"`
		Source       string `json:"source"`
		Notes        string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.BusinessName) == "" {
		http.Error(w, "business_name required", http.StatusBadRequest)
		return
	}
	p := &models.Prospect{
		BusinessName: strings.TrimSpace(body.BusinessName),
		Trade:        strings.TrimSpace(body.Trade),
		Location:     strings.TrimSpace(body.Location),
		Phone:        strings.TrimSpace(body.Phone),
		Email:        strings.TrimSpace(body.Email),
		Website:      strings.TrimSpace(body.Website),
		Source:       strings.TrimSpace(body.Source),
		Status:       "new",
		Notes:        strings.TrimSpace(body.Notes),
	}
	if err := h.store.CreateProspect(p); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"id":%d}`, p.ID)
}

func (h *Handler) AdminProspects(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("status")
	prospects, err := h.store.ListProspects(filter)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	h.render(w, "admin:prospects", map[string]any{
		"Prospects": prospects,
		"Filter":    filter,
		"Statuses":  prospectStatuses,
	})
}

func (h *Handler) AdminCreateProspect(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	p := &models.Prospect{
		BusinessName: strings.TrimSpace(r.FormValue("business_name")),
		Trade:        strings.TrimSpace(r.FormValue("trade")),
		Location:     strings.TrimSpace(r.FormValue("location")),
		Phone:        strings.TrimSpace(r.FormValue("phone")),
		Email:        strings.TrimSpace(r.FormValue("email")),
		Website:      strings.TrimSpace(r.FormValue("website")),
		Source:       strings.TrimSpace(r.FormValue("source")),
		Status:       "new",
		Notes:        strings.TrimSpace(r.FormValue("notes")),
	}
	if p.BusinessName == "" {
		http.Error(w, "business name required", http.StatusBadRequest)
		return
	}
	if err := h.store.CreateProspect(p); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/prospects", http.StatusSeeOther)
}

func (h *Handler) AdminProspect(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	p, err := h.store.GetProspectByID(id)
	if err != nil || p == nil {
		http.NotFound(w, r)
		return
	}
	h.render(w, "admin:prospect", map[string]any{
		"Prospect": p,
		"Statuses": prospectStatuses,
		"Updated":  r.URL.Query().Get("updated") == "1",
	})
}

func (h *Handler) AdminUpdateProspect(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	p, err := h.store.GetProspectByID(id)
	if err != nil || p == nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	p.BusinessName = strings.TrimSpace(r.FormValue("business_name"))
	p.Trade = strings.TrimSpace(r.FormValue("trade"))
	p.Location = strings.TrimSpace(r.FormValue("location"))
	p.Phone = strings.TrimSpace(r.FormValue("phone"))
	p.Email = strings.TrimSpace(r.FormValue("email"))
	p.Website = strings.TrimSpace(r.FormValue("website"))
	p.Source = strings.TrimSpace(r.FormValue("source"))
	p.Notes = strings.TrimSpace(r.FormValue("notes"))

	newStatus := r.FormValue("status")
	validStatus := false
	for _, s := range prospectStatuses {
		if s == newStatus {
			validStatus = true
			break
		}
	}
	if validStatus {
		// Set contacted_at when status first moves to contacted
		if newStatus == "contacted" && p.Status != "contacted" && p.ContactedAt == nil {
			now := time.Now().UTC()
			p.ContactedAt = &now
		}
		p.Status = newStatus
	}

	if err := h.store.UpdateProspect(p); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/prospects/%d?updated=1", id), http.StatusSeeOther)
}

func (h *Handler) AdminDeleteProspect(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := h.store.DeleteProspect(id); err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/prospects", http.StatusSeeOther)
}
