package handlers

import (
	"net/http"
	"strings"

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
