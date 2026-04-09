package handlers

import (
	"net/http"
	"strings"

	"github.com/adammcgrogan/locallaunch/internal/db"
	"github.com/adammcgrogan/locallaunch/internal/email"
	"github.com/adammcgrogan/locallaunch/internal/payment"
)

// siteTemplates lists available templates shown in the onboarding form.
// Industry is empty for general-purpose templates; set to the industry name for specific ones.
var siteTemplates = []struct {
	ID          string
	Name        string
	Description string
	ExampleSlug string
	Industry    string
}{
	{"bold", "Bold", "Dark, high contrast — great for trades and gyms", "example-bold", ""},
	{"fresh", "Fresh", "Light and minimal — ideal for professional services", "example-fresh", ""},
	{"warm", "Warm", "Earthy tones — perfect for cafés and restaurants", "example-warm", ""},
	{"glow", "Glow", "Soft pastels — suited for salons and beauty", "example-glow", ""},
	{"classic", "Classic", "Neutral and timeless — works for any business", "example-classic", ""},
	{"pulse", "Pulse", "Dark and energetic — built for gyms and fitness", "example-pulse", ""},
	{"grove", "Grove", "Forest green and organic — ideal for landscaping and garden", "example-grove", ""},
	{"fleet", "Fleet", "Industrial and direct — perfect for garages and auto services", "example-fleet", ""},
	{"haven", "Haven", "Warm and welcoming — great for B&Bs and holiday lets", "example-haven", ""},
	{"arch", "Arch", "Ultra-minimal and editorial — suited for design and creative services", "example-arch", ""},
	{"dine", "Dine", "Dark, moody layout with menu-style services section", "example-dine", "Restaurants & Food"},
	{"heal", "Heal", "Clean and clinical with trust badges in the hero", "example-heal", "Health & Wellness"},
	{"craft", "Craft", "Earthy and artisan with gallery as the centrepiece", "example-craft", "Makers & Studios"},
}

// buildTestimonials assembles the testimonials string from individual form fields.
func buildTestimonials(r *http.Request) string {
	names := r.Form["testimonial_name[]"]
	roles := r.Form["testimonial_role[]"]
	quotes := r.Form["testimonial_quote[]"]
	var lines []string
	for i, quote := range quotes {
		quote = strings.TrimSpace(quote)
		if quote == "" {
			continue
		}
		name := ""
		if i < len(names) {
			name = strings.TrimSpace(names[i])
		}
		role := ""
		if i < len(roles) {
			role = strings.TrimSpace(roles[i])
		}
		lines = append(lines, name+"|"+role+"|"+quote)
	}
	return strings.Join(lines, "\n")
}

// baseURL returns the scheme+host for the current request.
func (h *Handler) baseURL(reqHost string) string {
	scheme := "https"
	if strings.Contains(reqHost, ":") {
		scheme = "http"
	}
	return scheme + "://" + reqHost
}

// exampleURL builds the subdomain URL for an example site.
func (h *Handler) exampleURL(slug string) string {
	return "https://" + slug + "." + h.domain
}

// siteURL builds the public subdomain URL for a site.
func (h *Handler) siteURL(slug string) string {
	return "https://" + slug + "." + h.domain
}

type Handler struct {
	store          *db.Store
	email          *email.Client
	pay            *payment.Client
	domain         string
	adminPass      string
	umamiScriptURL string
}

func New(store *db.Store, email *email.Client, pay *payment.Client, domain, adminPass, umamiScriptURL string) *Handler {
	return &Handler{
		store:          store,
		email:          email,
		pay:            pay,
		domain:         domain,
		adminPass:      adminPass,
		umamiScriptURL: umamiScriptURL,
	}
}
