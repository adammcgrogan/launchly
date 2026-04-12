package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/adammcgrogan/launchly/internal/db"
	"github.com/adammcgrogan/launchly/internal/email"
	"github.com/adammcgrogan/launchly/internal/payment"
)

// siteTemplates lists available templates shown in the onboarding form.
// Industry is empty for general-purpose templates; set to the industry name for specific ones.
// Tags are short labels describing what kind of business each template suits.
var siteTemplates = []struct {
	ID          string
	Name        string
	Description string
	ExampleSlug string
	Industry    string
	Tags        []string
}{
	{"bold", "Bold", "Brutalist, industrial layout with safety-yellow accents and stamped labels", "example-bold", "", []string{"Builders", "Trades", "Construction", "Scaffolding", "Roofing"}},
	{"fresh", "Fresh", "Polished, modern layout with floating cards and soft gradients", "example-fresh", "", []string{"Accountants", "Consultants", "Solicitors", "Professional Services", "Agencies"}},
	{"warm", "Warm", "Rustic, handcrafted feel with serif display and letter-style storytelling", "example-warm", "", []string{"Cafés", "Bakeries", "Florists", "Small Shops", "Artisans"}},
	{"glow", "Glow", "Luxe, ornamental design with blush tones and flowing serif headlines", "example-glow", "", []string{"Salons", "Spas", "Beauty", "Nails", "Aesthetics"}},
	{"classic", "Classic", "Formal corporate layout with navy, gold accents and numbered sections", "example-classic", "", []string{"Solicitors", "Accountants", "Financial Advisors", "Estate Agents", "Consulting"}},
	{"pulse", "Pulse", "Aggressive, all-caps layout with electric-lime accents and oversized numbers", "example-pulse", "", []string{"Gyms", "Personal Trainers", "CrossFit", "Martial Arts", "Fitness Studios"}},
	{"grove", "Grove", "Organic split layout with forest green, warm cream and nature accents", "example-grove", "", []string{"Landscapers", "Gardeners", "Tree Surgeons", "Garden Design", "Grounds Maintenance"}},
	{"fleet", "Fleet", "Urgent, phone-first layout with hazard orange for call-out trades", "example-fleet", "", []string{"Mechanics", "MOT Centres", "Breakdown", "Taxi", "Removals"}},
	{"haven", "Haven", "Warm, hospitable layout with booking focus and review bars", "example-haven", "", []string{"B&Bs", "Holiday Lets", "Guesthouses", "Cottages", "Airbnb Hosts"}},
	{"arch", "Arch", "Editorial, minimal layout with serif typography and underline forms", "example-arch", "", []string{"Architects", "Interior Design", "Photographers", "Creative Studios", "Artists"}},
	{"dine", "Dine", "Fine-dining layout with centred serif headings and menu-style services", "example-dine", "Restaurants & Food", []string{"Restaurants", "Bistros", "Wine Bars", "Fine Dining", "Private Chefs"}},
	{"heal", "Heal", "Clean, clinical layout with trust badges and booking-focused contact", "example-heal", "Health & Wellness", []string{"Dentists", "Physios", "Chiropractors", "Private Clinics", "Therapists"}},
	{"craft", "Craft", "Earthy, artisan layout with masonry gallery and story-led about section", "example-craft", "Makers & Studios", []string{"Makers", "Ceramics", "Woodwork", "Jewellery", "Print Studios"}},
	{"shop", "Shop", "Clean, terracotta-accented retail layout with product grid and strong opening hours", "example-shop", "Retail & Shops", []string{"Gift Shops", "Boutiques", "Homeware", "Florists", "Delis", "Bookshops"}},
	{"vow", "Vow", "Elegant, serif-led wedding layout with prominent testimonials and enquiry focus", "example-vow", "Events & Weddings", []string{"Wedding Planners", "Event Stylists", "Venues", "Florists", "Photographers", "Celebrants"}},
}

// templateEntry is used to pass template metadata to public-facing pages.
type templateEntry struct {
	ID          string
	Name        string
	Description string
	ExampleURL  string
	Industry    string
	Tags        []string
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
	store             *db.Store
	email             *email.Client
	pay               *payment.Client
	domain            string
	adminPass         string
	umamiScriptURL    string
	tmpl              map[string]*template.Template
	contactLimiter    *rateLimiter // 5 submissions per IP per minute
	onboardingLimiter *rateLimiter // 3 submissions per IP per 5 minutes
}

func New(store *db.Store, email *email.Client, pay *payment.Client, domain, adminPass, umamiScriptURL string) (*Handler, error) {
	h := &Handler{
		store:             store,
		email:             email,
		pay:               pay,
		domain:            domain,
		adminPass:         adminPass,
		umamiScriptURL:    umamiScriptURL,
		tmpl:              make(map[string]*template.Template),
		contactLimiter:    newRateLimiter(1, 10*time.Minute),
		onboardingLimiter: newRateLimiter(1, 10*time.Minute),
	}
	if err := h.loadTemplates(); err != nil {
		return nil, fmt.Errorf("load templates: %w", err)
	}
	return h, nil
}

// loadTemplates parses all templates once at startup and caches them.
// This avoids repeated disk I/O and prevents template.Must panics at request time.
func (h *Handler) loadTemplates() error {
	pub := func(key, page string) error {
		t, err := template.ParseFiles(
			"web/templates/public/home_base.html",
			"web/templates/public/"+page+".html",
		)
		if err != nil {
			return fmt.Errorf("public/%s: %w", page, err)
		}
		h.tmpl[key] = t
		return nil
	}
	adm := func(key, page string) error {
		t, err := template.ParseFiles(
			"web/templates/admin/base.html",
			"web/templates/admin/"+page+".html",
		)
		if err != nil {
			return fmt.Errorf("admin/%s: %w", page, err)
		}
		h.tmpl["admin:"+key] = t
		return nil
	}

	for _, p := range []string{"home", "templates", "onboarding", "thankyou", "payment_success", "privacy", "terms"} {
		if err := pub(p, p); err != nil {
			return err
		}
	}
	for _, p := range []string{"dashboard", "site", "edit", "switch_template"} {
		if err := adm(p, p); err != nil {
			return err
		}
	}
	for _, st := range siteTemplates {
		t, err := template.ParseFiles(
			"web/templates/sites/base.html",
			"web/templates/sites/"+st.ID+".html",
		)
		if err != nil {
			return fmt.Errorf("site/%s: %w", st.ID, err)
		}
		h.tmpl["site:"+st.ID] = t
	}
	return nil
}

// render executes a pre-parsed template by key, writing the result to w.
func (h *Handler) render(w http.ResponseWriter, key string, data any) {
	tmpl, ok := h.tmpl[key]
	if !ok {
		http.Error(w, "template not found", http.StatusInternalServerError)
		log.Printf("render: unknown template key %q", key)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("render %s: %v", key, err)
	}
}

// HealthCheck returns 200 if the database is reachable, 503 otherwise.
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Ping(); err != nil {
		http.Error(w, "db unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
