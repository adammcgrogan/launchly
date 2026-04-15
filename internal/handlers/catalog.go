package handlers

import (
	"net/http"
	"strings"
)

// Palette is a named colour scheme for a template.
type Palette struct {
	ID   string
	Name string
	// CSS holds :root variable overrides, e.g. "--c-primary:#e11d48;--c-primary-fg:#fff;"
	// Empty string means "use the template's built-in defaults".
	CSS string
}

// HeadingFont is a font option for template headings.
type HeadingFont struct {
	ID   string
	Name string
	CSS  string // font-family value injected into headings
}

// HeadingFonts lists available heading font choices shown in the admin panel.
var HeadingFonts = []HeadingFont{
	{"sans", "Sans", "system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif"},
	{"serif", "Serif", "Georgia, 'Times New Roman', serif"},
	{"mono", "Mono", "'Courier New', Courier, monospace"},
}

// siteTemplates lists all available templates.
// Category: "general" (works for any business) or "specific" (industry-tailored).
// Tags: visual mood descriptors shown in the template picker.
// Palettes: curated colour schemes; first entry is the default (CSS="" uses template built-in colours).
var siteTemplates = []struct {
	ID          string
	Name        string
	Description string
	ExampleSlug string
	Category    string
	Industry    string // for specific templates
	Tags        []string
	Palettes    []Palette
}{
	// ── GENERAL ────────────────────────────────────────────────────────────────

	{
		ID: "default", Name: "Default",
		Description: "Clean, versatile layout that suits any business",
		ExampleSlug: "example-default", Category: "general",
		Tags: []string{"clean", "professional", "minimal", "light"},
		Palettes: []Palette{
			{"indigo", "Indigo", ""},
			{"emerald", "Emerald", "--c-primary:#059669;--c-primary-fg:#fff;--c-accent:#d1fae5;--c-accent-fg:#065f46"},
			{"slate", "Slate", "--c-primary:#475569;--c-primary-fg:#fff;--c-accent:#f1f5f9;--c-accent-fg:#1e293b"},
			{"crimson", "Crimson", "--c-primary:#e11d48;--c-primary-fg:#fff;--c-accent:#ffe4e6;--c-accent-fg:#9f1239"},
			{"dark", "Dark", "--c-bg:#0f172a;--c-surface:#1e293b;--c-primary:#6366f1;--c-primary-fg:#fff;--c-accent:#312e81;--c-accent-fg:#a5b4fc;--c-text:#94a3b8;--c-heading:#f1f5f9;--c-muted:#64748b;--c-border:#334155"},
		},
	},
	{
		ID: "canvas", Name: "Canvas",
		Description: "Modern light layout with floating cards and soft shadows",
		ExampleSlug: "example-canvas", Category: "general",
		Tags: []string{"light", "modern", "clean", "cards"},
		Palettes: []Palette{
			{"cobalt", "Cobalt", ""},
			{"violet", "Violet", "--c-primary:#7c3aed;--c-primary-fg:#fff;--c-accent:#ede9fe;--c-accent-fg:#6d28d9"},
			{"emerald", "Emerald", "--c-primary:#059669;--c-primary-fg:#fff;--c-accent:#d1fae5;--c-accent-fg:#065f46"},
			{"slate", "Slate", "--c-primary:#475569;--c-primary-fg:#fff;--c-accent:#f1f5f9;--c-accent-fg:#1e293b"},
		},
	},
	{
		ID: "slate", Name: "Slate",
		Description: "Dark, formal layout with strong typography and gold accents",
		ExampleSlug: "example-slate", Category: "general",
		Tags: []string{"dark", "formal", "refined", "strong"},
		Palettes: []Palette{
			{"gold", "Gold", ""},
			{"cyan", "Cyan", "--c-primary:#06b6d4;--c-primary-fg:#0f172a;--c-accent-fg:#67e8f9"},
			{"rose", "Rose", "--c-primary:#f43f5e;--c-primary-fg:#0f172a;--c-accent-fg:#fda4af"},
			{"lime", "Lime", "--c-primary:#84cc16;--c-primary-fg:#0f172a;--c-accent-fg:#bef264"},
		},
	},
	{
		ID: "hearth", Name: "Hearth",
		Description: "Warm, earthy layout with rustic serif and amber tones",
		ExampleSlug: "example-hearth", Category: "general",
		Tags: []string{"warm", "earthy", "cosy", "inviting"},
		Palettes: []Palette{
			{"amber", "Amber", ""},
			{"terracotta", "Terracotta", "--c-primary:#c2410c;--c-primary-fg:#fef3c7;--c-accent:#7c2d12"},
			{"sage", "Sage", "--c-primary:#4d7c0f;--c-primary-fg:#fef3c7;--c-accent:#3f6212"},
			{"navy", "Navy", "--c-primary:#1e40af;--c-primary-fg:#fef3c7;--c-accent:#1e3a8a"},
		},
	},
	{
		ID: "linen", Name: "Linen",
		Description: "Minimal, editorial layout with generous white space",
		ExampleSlug: "example-linen", Category: "general",
		Tags: []string{"minimal", "editorial", "elegant", "airy"},
		Palettes: []Palette{
			{"black", "Black", ""},
			{"indigo", "Indigo", "--c-primary:#4f46e5;--c-primary-fg:#fff;--c-accent:#e0e7ff;--c-accent-fg:#3730a3"},
			{"forest", "Forest", "--c-primary:#166534;--c-primary-fg:#fff;--c-accent:#dcfce7;--c-accent-fg:#14532d"},
			{"blush", "Blush", "--c-primary:#9f1239;--c-primary-fg:#fff;--c-accent:#ffe4e6;--c-accent-fg:#881337"},
		},
	},
	{
		ID: "onyx", Name: "Onyx",
		Description: "Very dark, high-contrast layout with bold typography",
		ExampleSlug: "example-onyx", Category: "general",
		Tags: []string{"dark", "bold", "striking", "impactful"},
		Palettes: []Palette{
			{"white", "White", ""},
			{"amber", "Amber", "--c-primary:#f59e0b;--c-primary-fg:#0a0a0a;--c-accent:#f59e0b;--c-accent-fg:#0a0a0a"},
			{"indigo", "Indigo", "--c-primary:#818cf8;--c-primary-fg:#0a0a0a;--c-accent:#818cf8;--c-accent-fg:#0a0a0a"},
			{"crimson", "Crimson", "--c-primary:#f87171;--c-primary-fg:#0a0a0a;--c-accent:#f87171;--c-accent-fg:#0a0a0a"},
		},
	},
	{
		ID: "ink", Name: "Ink",
		Description: "Deep navy layout with sky-blue accents, trustworthy and modern",
		ExampleSlug: "example-ink", Category: "general",
		Tags: []string{"navy", "trustworthy", "modern", "sharp"},
		Palettes: []Palette{
			{"sky", "Sky", ""},
			{"gold", "Gold", "--c-primary:#f59e0b;--c-primary-fg:#0c1a2e;--c-accent:#f59e0b;--c-accent-fg:#0c1a2e"},
			{"teal", "Teal", "--c-primary:#2dd4bf;--c-primary-fg:#0c1a2e;--c-accent:#2dd4bf;--c-accent-fg:#0c1a2e"},
			{"rose", "Rose", "--c-primary:#fb7185;--c-primary-fg:#0c1a2e;--c-accent:#fb7185;--c-accent-fg:#0c1a2e"},
		},
	},
	{
		ID: "copper", Name: "Copper",
		Description: "Warm white layout with terracotta primary, premium and grounded",
		ExampleSlug: "example-copper", Category: "general",
		Tags: []string{"warm", "terracotta", "premium", "grounded"},
		Palettes: []Palette{
			{"terracotta", "Terracotta", ""},
			{"sage", "Sage", "--c-primary:#4d7c0f;--c-primary-fg:#fff;--c-accent:#f0fdf4;--c-accent-fg:#3f6212;--c-heading:#14532d"},
			{"indigo", "Indigo", "--c-primary:#4f46e5;--c-primary-fg:#fff;--c-accent:#eef2ff;--c-accent-fg:#3730a3;--c-heading:#1e1b4b"},
			{"navy", "Navy", "--c-primary:#1e3a8a;--c-primary-fg:#fff;--c-accent:#dbeafe;--c-accent-fg:#1e40af;--c-heading:#0f1f5c"},
		},
	},

	// ── BUSINESS SPECIFIC ──────────────────────────────────────────────────────

	{
		ID: "builder", Name: "Builder",
		Description: "Brutalist dark layout with safety-yellow accents for trades",
		ExampleSlug: "example-builder", Category: "specific", Industry: "Trades & Construction",
		Tags: []string{"industrial", "dark", "gritty", "strong"},
		Palettes: []Palette{
			{"yellow", "Yellow", ""},
			{"orange", "Orange", "--c-primary:#f97316;--c-primary-fg:#0b0b0b;--c-accent-fg:#f97316"},
			{"red", "Red", "--c-primary:#ef4444;--c-primary-fg:#0b0b0b;--c-accent-fg:#ef4444"},
			{"blue", "Blue", "--c-primary:#3b82f6;--c-primary-fg:#0b0b0b;--c-accent-fg:#3b82f6"},
		},
	},
	{
		ID: "salon", Name: "Salon",
		Description: "Luxe, ornamental design with blush tones for beauty businesses",
		ExampleSlug: "example-salon", Category: "specific", Industry: "Beauty & Wellness",
		Tags: []string{"soft", "feminine", "luxe", "rose"},
		Palettes: []Palette{
			{"rose", "Rose", ""},
			{"violet", "Violet", "--c-primary:#7c3aed;--c-primary-fg:#fff;--c-accent:#fdf4ff;--c-accent-fg:#6b21a8"},
			{"gold", "Gold", "--c-primary:#b45309;--c-primary-fg:#fff;--c-accent:#fff9f0;--c-accent-fg:#92400e"},
			{"teal", "Teal", "--c-primary:#0d9488;--c-primary-fg:#fff;--c-accent:#f0fdfa;--c-accent-fg:#0f766e"},
		},
	},
	{
		ID: "gym", Name: "Gym",
		Description: "Aggressive all-caps layout with electric accents for fitness",
		ExampleSlug: "example-gym", Category: "specific", Industry: "Fitness & Sport",
		Tags: []string{"energetic", "electric", "dark", "bold"},
		Palettes: []Palette{
			{"lime", "Lime", ""},
			{"electric", "Electric", "--c-primary:#00d4ff;--c-primary-fg:#0a0a0a;--c-accent-fg:#00d4ff"},
			{"orange", "Orange", "--c-primary:#ff6b00;--c-primary-fg:#0a0a0a;--c-accent-fg:#ff6b00"},
			{"red", "Red", "--c-primary:#ff2020;--c-primary-fg:#0a0a0a;--c-accent-fg:#ff2020"},
		},
	},
	{
		ID: "landscaping", Name: "Landscaping",
		Description: "Organic split layout with forest green and warm cream tones",
		ExampleSlug: "example-landscaping", Category: "specific", Industry: "Landscaping & Gardens",
		Tags: []string{"natural", "green", "organic", "earthy"},
		Palettes: []Palette{
			{"forest", "Forest", ""},
			{"sage", "Sage", "--c-primary:#84cc16;--c-primary-fg:#1a3a28;--c-accent-fg:#84cc16"},
			{"stone", "Stone", "--c-primary:#e7c5a0;--c-primary-fg:#1a3a28;--c-accent-fg:#e7c5a0"},
			{"bronze", "Bronze", "--c-primary:#d97706;--c-primary-fg:#1a3a28;--c-accent-fg:#d97706"},
		},
	},
	{
		ID: "garage", Name: "Garage",
		Description: "Urgent, phone-first layout with hazard orange for auto trades",
		ExampleSlug: "example-garage", Category: "specific", Industry: "Automotive",
		Tags: []string{"dark", "orange", "urgent", "mechanical"},
		Palettes: []Palette{
			{"orange", "Orange", ""},
			{"red", "Red", "--c-primary:#dc2626;--c-primary-fg:#fff;--c-accent-fg:#dc2626"},
			{"blue", "Blue", "--c-primary:#2563eb;--c-primary-fg:#fff;--c-accent-fg:#2563eb"},
			{"yellow", "Yellow", "--c-primary:#eab308;--c-primary-fg:#111;--c-accent-fg:#eab308"},
		},
	},
	{
		ID: "bnb", Name: "B&B",
		Description: "Warm, hospitable layout with booking focus for accommodation",
		ExampleSlug: "example-bnb", Category: "specific", Industry: "Accommodation",
		Tags: []string{"warm", "hospitable", "teal", "soft"},
		Palettes: []Palette{
			{"teal", "Teal", ""},
			{"sage", "Sage", "--c-primary:#4d7c0f;--c-primary-fg:#fff;--c-accent:#f0fdf4;--c-accent-fg:#3f6212"},
			{"indigo", "Indigo", "--c-primary:#4f46e5;--c-primary-fg:#fff;--c-accent:#eef2ff;--c-accent-fg:#3730a3"},
			{"amber", "Amber", "--c-primary:#d97706;--c-primary-fg:#fff;--c-accent:#fffbeb;--c-accent-fg:#92400e"},
		},
	},
	{
		ID: "restaurant", Name: "Restaurant",
		Description: "Fine-dining layout with dark moody tones and menu-style services",
		ExampleSlug: "example-restaurant", Category: "specific", Industry: "Food & Dining",
		Tags: []string{"dark", "rich", "moody", "elegant"},
		Palettes: []Palette{
			{"gold", "Gold", ""},
			{"crimson", "Crimson", "--c-primary:#dc2626;--c-primary-fg:#1a1110;--c-accent-fg:#dc2626"},
			{"sage", "Sage", "--c-primary:#4ade80;--c-primary-fg:#1a1110;--c-accent-fg:#4ade80"},
			{"sky", "Sky", "--c-primary:#38bdf8;--c-primary-fg:#1a1110;--c-accent-fg:#38bdf8"},
		},
	},
	{
		ID: "clinic", Name: "Clinic",
		Description: "Clean, clinical layout with trust badges for health businesses",
		ExampleSlug: "example-clinic", Category: "specific", Industry: "Health & Medical",
		Tags: []string{"clean", "clinical", "calm", "trustworthy"},
		Palettes: []Palette{
			{"teal", "Teal", ""},
			{"blue", "Blue", "--c-primary:#2563eb;--c-primary-fg:#fff;--c-accent:#dbeafe;--c-accent-fg:#1d4ed8"},
			{"violet", "Violet", "--c-primary:#7c3aed;--c-primary-fg:#fff;--c-accent:#ede9fe;--c-accent-fg:#6d28d9"},
			{"forest", "Forest", "--c-primary:#166534;--c-primary-fg:#fff;--c-accent:#dcfce7;--c-accent-fg:#14532d"},
		},
	},
	{
		ID: "maker", Name: "Maker",
		Description: "Earthy, artisan layout with textured feel for studios and crafts",
		ExampleSlug: "example-maker", Category: "specific", Industry: "Makers & Artisans",
		Tags: []string{"earthy", "textured", "handmade", "warm"},
		Palettes: []Palette{
			{"clay", "Clay", ""},
			{"forest", "Forest", "--c-primary:#3f6212;--c-primary-fg:#faf6f1;--c-accent:#1a3a28"},
			{"indigo", "Indigo", "--c-primary:#4f46e5;--c-primary-fg:#faf6f1;--c-accent:#312e81"},
			{"rust", "Rust", "--c-primary:#b45309;--c-primary-fg:#faf6f1;--c-accent:#78350f"},
		},
	},
	{
		ID: "retail", Name: "Retail",
		Description: "Bright terracotta-accented layout for shops and boutiques",
		ExampleSlug: "example-retail", Category: "specific", Industry: "Retail & Shops",
		Tags: []string{"bright", "fresh", "terracotta", "welcoming"},
		Palettes: []Palette{
			{"terracotta", "Terracotta", ""},
			{"sage", "Sage", "--c-primary:#4d7c0f;--c-primary-fg:#fff;--c-accent:#f0fdf4;--c-accent-fg:#3f6212"},
			{"navy", "Navy", "--c-primary:#1e3a8a;--c-primary-fg:#fff;--c-accent:#dbeafe;--c-accent-fg:#1e40af"},
			{"berry", "Berry", "--c-primary:#7e22ce;--c-primary-fg:#fff;--c-accent:#faf5ff;--c-accent-fg:#6b21a8"},
		},
	},
	{
		ID: "wedding", Name: "Wedding",
		Description: "Elegant serif-led layout for weddings and events",
		ExampleSlug: "example-wedding", Category: "specific", Industry: "Weddings & Events",
		Tags: []string{"elegant", "romantic", "soft", "serif"},
		Palettes: []Palette{
			{"classic", "Classic", ""},
			{"blush", "Blush", "--c-primary:#9f1239;--c-primary-fg:#fdf9f6;--c-accent:#fff1f2;--c-accent-fg:#9f1239"},
			{"forest", "Forest", "--c-primary:#166534;--c-primary-fg:#fdf9f6;--c-accent:#f0fdf4;--c-accent-fg:#166534"},
			{"gold", "Gold", "--c-primary:#92400e;--c-primary-fg:#fdf9f6;--c-accent:#fef3c7;--c-accent-fg:#78350f"},
		},
	},
	{
		ID: "barber", Name: "Barber",
		Description: "Sharp, masculine dark layout for barbershops and men's grooming",
		ExampleSlug: "example-barber", Category: "specific", Industry: "Barbering & Grooming",
		Tags: []string{"dark", "sharp", "masculine", "slick"},
		Palettes: []Palette{
			{"gold", "Gold", ""},
			{"red", "Red", "--c-primary:#dc2626;--c-primary-fg:#0a0a0a;--c-accent:#dc2626"},
			{"blue", "Blue", "--c-primary:#2563eb;--c-primary-fg:#0a0a0a;--c-accent:#2563eb"},
			{"green", "Green", "--c-primary:#16a34a;--c-primary-fg:#0a0a0a;--c-accent:#16a34a"},
		},
	},
	{
		ID: "takeaway", Name: "Takeaway",
		Description: "Bold, vibrant layout built for takeaways and fast food",
		ExampleSlug: "example-takeaway", Category: "specific", Industry: "Takeaway & Fast Food",
		Tags: []string{"bold", "vibrant", "casual", "energetic"},
		Palettes: []Palette{
			{"orange", "Orange", ""},
			{"red", "Red", "--c-primary:#dc2626;--c-primary-fg:#fff;--c-accent:#fee2e2;--c-accent-fg:#991b1b"},
			{"green", "Green", "--c-primary:#16a34a;--c-primary-fg:#fff;--c-accent:#dcfce7;--c-accent-fg:#166534"},
			{"blue", "Blue", "--c-primary:#2563eb;--c-primary-fg:#fff;--c-accent:#dbeafe;--c-accent-fg:#1d4ed8"},
		},
	},
}

// templateEntry is used to pass template metadata to public-facing pages.
type templateEntry struct {
	ID          string
	Name        string
	Description string
	ExampleURL  string
	Category    string
	Industry    string
	Tags        []string
}

// getPaletteCSS returns the CSS variable string for the given template + palette combination.
// Returns "" if not found (template uses its own defaults).
func getPaletteCSS(templateID, paletteID string) string {
	for _, t := range siteTemplates {
		if t.ID == templateID {
			for _, p := range t.Palettes {
				if p.ID == paletteID {
					return p.CSS
				}
			}
			return ""
		}
	}
	return ""
}

// getHeadingFontCSS returns the font-family CSS value for the given font ID.
func getHeadingFontCSS(fontID string) string {
	for _, f := range HeadingFonts {
		if f.ID == fontID {
			return f.CSS
		}
	}
	return ""
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
