package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/adammcgrogan/launchly/internal/models"
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

// SeedExamples inserts or updates pre-published example sites for each template.
func (s *Store) SeedExamples() error {
	examples := []models.Site{
		{
			Slug: "example-bold", BusinessName: "McLaughlin Plumbing & Heating", Template: "bold",
			CTAText:        "Get a Quote",
			Tagline:        "Belfast's most trusted plumbers — available 24/7",
			About:          "Family-run plumbing and heating business proudly serving Belfast and Greater Northern Ireland since 1998. Gas Safe registered, fully insured, and on call around the clock for emergencies.",
			Services:       "Emergency Call-Out — 24/7\nBoiler Repair & Servicing\nLeak Detection & Repair\nBathroom & Wet Room Fitting\nCentral Heating Installation\nLandlord Gas Safety Certificates",
			Certifications: "24/7 Emergency Callout\n25+ Years Experience\nGas Safe Registered\n★★★★★ Rated Locally",
			Location:       "Belfast, NI",
			Phone:          "028 9011 2233",
			Email:          "info@mclaughlinplumbing.co.uk",
			Address:        "14 Donegall Road, Belfast, BT12 5JN",
			Hours:          "Mon–Fri: 7am – 7pm\nSaturday: 8am – 4pm\nEmergency: 24/7",
			LeadEmail:      "example@launchly.ltd",
		},
		{
			Slug: "example-fresh", BusinessName: "O'Neill Accountancy", Template: "fresh",
			CTAText:        "Book a Consultation",
			Tagline:        "Plain-talking accountants for Northern Ireland businesses",
			About:          "O'Neill Accountancy has been keeping the books straight for sole traders and SMEs across Northern Ireland since 2008. We cut through the jargon and give you advice that actually makes a difference.",
			Services:       "Self-Assessment Tax Returns\nPayroll Management\nBookkeeping & VAT Returns\nBusiness Start-Up Advice\nYear-End Accounts\nR&D Tax Credits",
			Certifications: "ACCA Qualified\nICB Registered\nFree Initial Consultation\n15+ Years Experience",
			Location:       "Derry, NI",
			Phone:          "028 7134 5678",
			Email:          "hello@oneillaccountancy.co.uk",
			Address:        "Unit 3, Ebrington Square, Derry, BT47 6FA",
			Hours:          "Mon–Fri: 9am – 5:30pm\nSaturday: By appointment",
			LeadEmail:      "example@launchly.ltd",
		},
		{
			Slug: "example-warm", BusinessName: "The Wee Bakehouse", Template: "warm",
			CTAText:        "Find Us",
			Tagline:        "Freshly baked every morning in the heart of Lisburn",
			About:          "A proper local bakery and café baking everything from scratch since 2011. We source our flour from a mill in Co. Antrim and our eggs from a farm just up the road. Come in, sit down, and enjoy something homemade.",
			Services:       "Ulster Fry — the full works\nFreshly Baked Soda & Wheaten Bread\nHomemade Soups & Toasties\nCakes, Traybakes & Scones\nCoffee & Teas\nWhole Cakes to Order",
			Certifications: "Baked Fresh Daily\nLocal Ingredients\nFamily Run Since 2011\nDine In & Takeaway",
			Location:       "Lisburn, NI",
			Phone:          "028 9266 7788",
			Email:          "hello@theweebakehouse.co.uk",
			Address:        "22 Market Square, Lisburn, BT28 1AG",
			Hours:          "Mon–Sat: 7:30am – 4pm\nSunday: 9am – 2pm",
			LeadEmail:      "example@launchly.ltd",
		},
		{
			Slug: "example-glow", BusinessName: "Aoife's Beauty Studio", Template: "glow",
			CTAText:        "Book Now",
			Tagline:        "Award-winning hair & beauty in the heart of Newry",
			About:          "Aoife's Beauty Studio has been making clients look and feel amazing since 2015. From colour and cuts to lashes and nails — our fully qualified team use only premium products for results that last.",
			Services:       "Cut & Blow Dry\nColour, Highlights & Balayage\nLash Extensions & Lifts\nGel Nails & Manicures\nBridal Hair & Beauty\nKeratin Smoothing Treatments",
			Certifications: "Award Winning Studio\nFully Qualified Team\nPremium Products Only\nBooking Essential",
			Location:       "Newry, NI",
			Phone:          "028 3026 1122",
			Email:          "book@aoifesbeauty.co.uk",
			Address:        "8 Hill Street, Newry, BT34 1AR",
			Hours:          "Tue–Fri: 9am – 7pm\nSaturday: 9am – 5pm\nSun & Mon: Closed",
			LeadEmail:      "example@launchly.ltd",
		},
		{
			Slug: "example-classic", BusinessName: "Quinn Electrical Services", Template: "classic",
			CTAText:        "Get a Free Quote",
			Tagline:        "NICEIC approved electricians serving Co. Antrim",
			About:          "Quinn Electrical Services delivers safe, reliable domestic and commercial electrical work across Co. Antrim and beyond. Every job is fully tested, certified, and completed to Part P building regulations.",
			Services:       "Full House Rewiring\nConsumer Unit Upgrades\nLighting Design & Installation\nEV Charger Installation\nSmart Home & Security Systems\nPAT Testing",
			Certifications: "NICEIC Approved\nPart P Certified\nFully Insured\nFree Quotations",
			Location:       "Co. Antrim, NI",
			Phone:          "028 9443 5566",
			Email:          "info@quinnelectrical.co.uk",
			Address:        "17 Railway Street, Antrim, BT41 4AE",
			Hours:          "Mon–Fri: 7:30am – 6pm\nSaturday: 8am – 1pm",
			LeadEmail:      "example@launchly.ltd",
		},
		{
			Slug: "example-pulse", BusinessName: "Titan Fitness Belfast", Template: "pulse",
			CTAText:        "Join Now",
			Tagline:        "Belfast's hardest-working gym — no excuses",
			About:          "Titan Fitness is a serious training facility in the heart of Belfast. No fluff, no gimmicks — just quality equipment, expert coaching, and a community that shows up. Whether you're a first-timer or a seasoned lifter, we'll push you further.",
			Services:       "Strength & Conditioning\nGroup HIIT Classes\n1-on-1 Personal Training\nNutrition Coaching\nBoxing & Kickboxing\nYoga & Mobility",
			Certifications: "Open 6am – 10pm Daily\nQualified PTs on Floor\nNo Contract Membership\nFree 1-Week Trial",
			Location:       "Belfast, NI",
			Phone:          "028 9031 4488",
			Email:          "hello@titanfitnessbelfast.co.uk",
			Address:        "Unit 5, Boucher Road Industrial Estate, Belfast, BT12 6HR",
			Hours:          "Mon–Fri: 6am – 10pm\nSaturday: 7am – 8pm\nSunday: 8am – 6pm",
			LeadEmail:      "example@launchly.ltd",
		},
		{
			Slug: "example-grove", BusinessName: "O'Hara Landscaping", Template: "grove",
			CTAText:        "Get a Free Quote",
			Tagline:        "Transforming gardens across Co. Down since 2007",
			About:          "O'Hara Landscaping designs and builds beautiful outdoor spaces for homes and businesses across Co. Down and beyond. From a simple lawn makeover to a full garden redesign — we take pride in every square foot.",
			Services:       "Garden Design & Planning\nLawn Installation & Maintenance\nDecking & Patio Construction\nPlanting & Borders\nFencing & Boundary Work\nIrrigation Systems",
			Certifications: "Fully Insured\nFree Site Visit\nOver 15 Years Experience\nAll Work Guaranteed",
			Location:       "Downpatrick, Co. Down",
			Phone:          "028 4461 2277",
			Email:          "info@oharalandscaping.co.uk",
			Address:        "The Yard, Strangford Road, Downpatrick, BT30 6JT",
			Hours:          "Mon–Fri: 8am – 5:30pm\nSaturday: 9am – 1pm",
			LeadEmail:      "example@launchly.ltd",
		},
		{
			Slug: "example-fleet", BusinessName: "Brady's Auto Centre", Template: "fleet",
			CTAText:        "Book a Service",
			Tagline:        "MOT, servicing & repairs you can trust in Armagh",
			About:          "Brady's Auto Centre has been keeping Armagh on the road since 1994. We're an authorised MOT test centre with a fully equipped workshop handling everything from a quick tyre swap to a full engine rebuild. Honest prices, no surprises.",
			Services:       "MOT Testing (Classes 1–4)\nFull & Interim Car Servicing\nBrakes, Clutch & Exhaust\nTyres — Supply & Fit\nAir Conditioning Regas\nDiagnostics & Fault Finding",
			Certifications: "DVA Authorised MOT Centre\n30 Years in Business\nAll Makes & Models\nFree Courtesy Car",
			Location:       "Armagh, NI",
			Phone:          "028 3752 1199",
			Email:          "bookings@bradysauto.co.uk",
			Address:        "45 Lonsdale Road, Armagh, BT61 7HZ",
			Hours:          "Mon–Fri: 8am – 6pm\nSaturday: 8:30am – 1pm",
			LeadEmail:      "example@launchly.ltd",
		},
		{
			Slug: "example-haven", BusinessName: "The Harbour House", Template: "haven",
			CTAText:        "Check Availability",
			Tagline:        "A peaceful waterfront retreat in Strangford, Co. Down",
			About:          "The Harbour House is a beautifully restored Victorian townhouse overlooking Strangford Lough. Offering five en-suite rooms, a guest lounge, and a hearty homemade breakfast each morning — it's the perfect base for exploring Co. Down.",
			Services:       "En-suite Double & Twin Rooms\nHomemade Full Irish Breakfast\nEarly Check-in on Request\nFree Private Parking\nCycle Storage & Drying Room\nLocal Walking Routes & Maps",
			Certifications: "Tourism NI Approved\nTripadvisor Certificate of Excellence\nFree Cancellation Policy\nFamily & Pet Friendly",
			Location:       "Strangford, Co. Down",
			Phone:          "028 4488 1556",
			Email:          "stay@theharbourhouse.co.uk",
			Address:        "2 The Quay, Strangford, BT30 7NF",
			Hours:          "Check-in: 3pm – 9pm\nCheck-out: by 11am\nBreakfast: 7:30am – 9:30am",
			LeadEmail:      "example@launchly.ltd",
		},
		{
			Slug: "example-arch", BusinessName: "Laura Vance Interiors", Template: "arch",
			CTAText:        "Start a Project",
			Tagline:        "Considered interior design for homes and businesses",
			About:          "Laura Vance Interiors is a Belfast-based studio specialising in residential and commercial interior design. Every project begins with listening — understanding how you live, work, and what you want a space to feel like. The result is always intentional, always personal.",
			Services:       "Full Interior Design\nSpace Planning & Layouts\nFurniture Sourcing & Styling\nColour & Material Consultancy\nKitchen & Bathroom Design\nCommercial & Office Interiors",
			Certifications: "BIID Affiliated Designer\nFully Insured\nFree Initial Consultation\nNationwide Projects",
			Location:       "Belfast, NI",
			Phone:          "028 9024 3311",
			Email:          "studio@lauravanceinteriors.co.uk",
			Address:        "Studio 12, Cathedral Quarter, Belfast, BT1 1FB",
			Hours:          "Mon–Fri: 9am – 6pm\nSaturday: By appointment",
			LeadEmail:      "example@launchly.ltd",
		},
		{
			Slug: "example-dine", BusinessName: "The Ember Room", Template: "dine",
			CTAText:        "Book a Table",
			Tagline:        "Wood-fired food and natural wine in the heart of Belfast",
			About:          "The Ember Room opened in 2019 in a converted Victorian warehouse on Hill Street. Everything we cook passes through the wood-fired oven or over the grill — we believe fire makes food taste better. Our menu changes weekly to follow what's seasonal and local.",
			Services:       "Wood-Fired Sourdough Pizza\nCharcuterie & Sharing Boards\nGrilled Fish of the Day\nRump Cap & Short Rib\nNatural & Organic Wine List\nSunday Set Menu — 3 courses",
			Certifications: "Booking Recommended\nPrivate Dining Available\nSourced Within 50 Miles\nOpen Kitchen",
			Testimonials:   "Ciara Murphy|Regular|Best pizza I've had outside of Naples. The dough is incredible and the toppings are always something a bit different. Our go-to for date night.\nJames Devlin|Food Blogger|The short rib with chimichurri was genuinely one of the best things I've eaten in Belfast this year. The natural wine list is thoughtful too.\nSaoirse & Tom||Booked out the private dining room for our anniversary and the team were exceptional. Food, wine, and atmosphere — all perfect.",
			Location:       "Belfast, NI",
			Phone:          "028 9031 7744",
			Email:          "hello@theemberroom.co.uk",
			Address:        "14 Hill Street, Cathedral Quarter, Belfast, BT1 2LB",
			Hours:          "Wed–Thu: 5pm – 10pm\nFri–Sat: 12pm – 11pm\nSunday: 1pm – 8pm\nMon & Tue: Closed",
			LeadEmail:      "example@launchly.ltd",
		},
		{
			Slug: "example-heal", BusinessName: "Greenfield Dental", Template: "heal",
			CTAText:        "Book an Appointment",
			Tagline:        "Gentle, modern dentistry for the whole family in Lisburn",
			About:          "Greenfield Dental has been caring for smiles in Lisburn since 2004. Our team of four dentists and three hygienists provide a full range of NHS and private treatments in a calm, unhurried environment. We see patients of all ages — from toddlers to grandparents.",
			Services:       "NHS & Private Check-Ups\nHygiene & Scale & Polish\nTooth Whitening\nInvisalign & Clear Aligners\nDental Implants\nComposite Bonding\nEmergency Appointments",
			Certifications: "GDC Registered Practitioners\nNHS & Private Patients Welcome\nSame-Day Emergency Slots\nInterest-Free Payment Plans",
			Testimonials:   "Patricia Hagan|Patient since 2011|I used to dread the dentist. The team at Greenfield are so gentle and patient — I actually look forward to my check-ups now. Couldn't recommend them more.\nDr. Michael Corrigan|GP Referral|I regularly refer patients to Greenfield. The standard of care is excellent and my patients always come back having had a positive experience.\nEmma & Patrick Walsh||Brought our three kids here for the first time and the staff were brilliant with them. Our youngest even asked when she can come back!",
			Location:       "Lisburn, NI",
			Phone:          "028 9266 4433",
			Email:          "reception@greenfielddental.co.uk",
			Address:        "7 Governors Road, Lisburn, BT28 1EL",
			Hours:          "Mon–Thu: 8:30am – 5:30pm\nFriday: 8:30am – 4pm\nSaturday: 9am – 1pm (Private)\nSunday: Closed",
			LeadEmail:      "example@launchly.ltd",
		},
		{
			Slug: "example-shop", BusinessName: "The Corner Collective", Template: "shop",
			CTAText:        "Visit Us",
			Tagline:        "Thoughtfully chosen gifts, homewares and local finds",
			About:          "The Corner Collective has been a fixture of the Lisburn Road since 2016. We stock an ever-changing mix of gifts, ceramics, candles and homewares — most of it sourced from small Irish and British makers. Pop in and have a browse, you'll always find something worth taking home.",
			Services:       "Gifts & Keepsakes\nHomeware & Ceramics\nCandles & Fragrance\nCards & Stationery\nLocal Maker Collection\nCorporate & Bulk Gifting",
			Certifications: "Local Makers Stocked\nGift Wrapping Available\nClick & Collect\nOpen 7 Days",
			Testimonials:   "Claire Donnelly||The most beautiful little shop — I always end up spending way more than I planned. The staff are so helpful and the gift wrapping is gorgeous.\nMark & Louise Forde|Regular Customers|We've bought almost every birthday and Christmas present here for the last three years. There's always something new in and the quality is brilliant.\nSophie McAuley||Found the perfect wedding gift here that I couldn't find anywhere else. The owner took time to help me and even added a handwritten note.",
			Location:       "Lisburn Road, Belfast",
			Phone:          "028 9066 1234",
			Email:          "hello@cornercollective.co.uk",
			Address:        "142 Lisburn Road, Belfast, BT9 6AJ",
			Hours:          "Mon–Sat: 9:30am – 5:30pm\nSunday: 12pm – 4pm",
			LeadEmail:      "example@launchly.ltd",
		},
		{
			Slug: "example-vow", BusinessName: "Clover & White Events", Template: "vow",
			CTAText:        "Get in Touch",
			Tagline:        "Beautifully planned weddings and events across Ireland",
			About:          "Clover & White is a wedding and events planning studio based in Co. Down. Founded by Aoife Connolly in 2018, we believe every wedding should feel entirely personal — a true reflection of the people at the heart of it. We take care of every detail so you can be present for every moment.",
			Services:       "Full Wedding Planning\nPartial Planning & Support\nDay-of Coordination\nVenue Styling & Décor\nFloral Design & Arrangements\nElopements & Intimate Weddings",
			Certifications: "Fully Insured\nFree Initial Consultation\nIreland & UK Wide\nAvailable Weekends",
			Testimonials:   "Niamh & Ciarán Kelly||We genuinely couldn't have done it without Aoife. She thought of things we'd never have considered and kept everything calm on the day. Our wedding was absolutely perfect.\nEmma Doherty|Maid of Honour|As the maid of honour I was dreading all the logistics. Clover & White handled everything — the venue looked stunning and the whole day ran like clockwork.\nPaul & Sarah McBride||From our first call to the last dance, Aoife was warm, professional and completely on top of everything. Worth every penny and more.",
			Location:       "Co. Down, Ireland",
			Phone:          "077 9900 1122",
			Email:          "hello@cloverandwhite.co.uk",
			Address:        "Downpatrick, Co. Down, BT30",
			Hours:          "Consultations by appointment\nMon–Fri: 9am – 6pm\nWeekends: Available for events",
			LeadEmail:      "example@launchly.ltd",
		},
		{
			Slug: "example-craft", BusinessName: "Willow & Thread", Template: "craft",
			CTAText:        "View Collection",
			Tagline:        "Handthrown ceramics and homeware made in Co. Antrim",
			About:          "Willow & Thread is a one-woman ceramics studio run by Niamh Doyle from a converted outhouse on her family's farm in Ballymena. Every piece is handthrown on the wheel, glazed with natural oxides, and fired in a small kiln. No two are the same.",
			Services:       "Handthrown Mugs & Bowls\nVases & Bud Vases\nPlates & Serving Platters\nCustom Wedding & Gift Sets\nCorporate Gifting\nWheel-Throwing Workshops",
			Certifications: "Made by Hand in Co. Antrim\nNatural Glazes Only\nFood & Dishwasher Safe\nMade-to-Order Available",
			Testimonials:   "Rachel Clarke||Ordered a set of six mugs as a wedding gift and they were absolutely beautiful. Niamh was so helpful with customising the glaze colours. They arrived beautifully packaged too.\nTom & Dee McAllister|Workshop Participants|Did the wheel-throwing workshop as a date night — it was brilliant craic. Niamh is a great teacher and really patient. We each went home with a little wonky bowl we're very proud of.\nFiona Brennan|Interior Stylist|I've been sourcing Willow & Thread pieces for client projects for two years. The quality is consistently exceptional and Niamh always delivers on time.",
			Location:       "Ballymena, Co. Antrim",
			Phone:          "077 1234 9988",
			Email:          "hello@willowandthread.co.uk",
			Address:        "The Studio, Tullygarley Road, Ballymena, BT42 2QP",
			Hours:          "Studio visits by appointment\nOnline shop: open 24/7\nWorkshops: Fri & Sat evenings",
			LeadEmail:      "example@launchly.ltd",
		},
	}

	for _, e := range examples {
		existing, err := s.GetSiteBySlug(e.Slug)
		if err != nil {
			return err
		}
		site := e
		if existing != nil {
			site.ID = existing.ID
			if err := s.updateExampleSite(&site); err != nil {
				return err
			}
		} else {
			if err := s.CreateSite(&site); err != nil {
				return err
			}
			if err := s.PublishSite(site.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) updateExampleSite(site *models.Site) error {
	_, err := s.db.Exec(`
		UPDATE sites SET business_name=$1, tagline=$2, about=$3, services=$4,
		certifications=$5, location=$6, cta_text=$7, phone=$8, email=$9, address=$10,
		hours=$11, status='live' WHERE id=$12`,
		site.BusinessName, site.Tagline, site.About, site.Services,
		site.Certifications, site.Location, site.CTAText, site.Phone, site.Email,
		site.Address, site.Hours, site.ID,
	)
	return err
}

func (s *Store) UpdateSiteTemplate(id int, template string) error {
	_, err := s.db.Exec(`UPDATE sites SET template=$1 WHERE id=$2`, template, id)
	return err
}

// --- Sites ---

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

func (s *Store) UpdateSiteNotes(id int, notes string) error {
	_, err := s.db.Exec(`UPDATE sites SET notes = $1 WHERE id = $2`, notes, id)
	return err
}

func (s *Store) GetSiteBySlug(slug string) (*models.Site, error) {
	site := &models.Site{}
	err := s.db.QueryRow(`
		SELECT id, slug, business_name, template, tagline, about, services,
		       certifications, location, cta_text, testimonials, logo_url, gallery,
		       phone, email, address, hours, map_url, map_embed_url,
		       facebook_url, instagram_url, whatsapp_url, twitter_url, tiktok_url, linkedin_url, youtube_url,
		       umami_website_id, lead_email, status, created_at, published_at,
		       plan, payment_status, stripe_session_id, stripe_subscription_id, paid_at,
		       COALESCE(custom_domain, ''), notes, analytics_frequency, analytics_last_sent
		FROM sites WHERE slug = $1`, slug).
		Scan(&site.ID, &site.Slug, &site.BusinessName, &site.Template,
			&site.Tagline, &site.About, &site.Services,
			&site.Certifications, &site.Location, &site.CTAText,
			&site.Testimonials, &site.LogoURL, &site.Gallery,
			&site.Phone, &site.Email, &site.Address, &site.Hours, &site.MapURL, &site.MapEmbedURL,
			&site.FacebookURL, &site.InstagramURL, &site.WhatsAppURL,
			&site.TwitterURL, &site.TikTokURL, &site.LinkedInURL, &site.YouTubeURL,
			&site.UmamiWebsiteID, &site.LeadEmail, &site.Status, &site.CreatedAt, &site.PublishedAt,
			&site.Plan, &site.PaymentStatus, &site.StripeSessionID, &site.StripeSubscriptionID, &site.PaidAt,
			&site.CustomDomain, &site.Notes, &site.AnalyticsFrequency, &site.AnalyticsLastSent)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return site, err
}

func (s *Store) GetSiteByID(id int) (*models.Site, error) {
	site := &models.Site{}
	err := s.db.QueryRow(`
		SELECT id, slug, business_name, template, tagline, about, services,
		       certifications, location, cta_text, testimonials, logo_url, gallery,
		       phone, email, address, hours, map_url, map_embed_url,
		       facebook_url, instagram_url, whatsapp_url, twitter_url, tiktok_url, linkedin_url, youtube_url,
		       umami_website_id, lead_email, status, created_at, published_at,
		       plan, payment_status, stripe_session_id, stripe_subscription_id, paid_at,
		       COALESCE(custom_domain, ''), notes, analytics_frequency, analytics_last_sent
		FROM sites WHERE id = $1`, id).
		Scan(&site.ID, &site.Slug, &site.BusinessName, &site.Template,
			&site.Tagline, &site.About, &site.Services,
			&site.Certifications, &site.Location, &site.CTAText,
			&site.Testimonials, &site.LogoURL, &site.Gallery,
			&site.Phone, &site.Email, &site.Address, &site.Hours, &site.MapURL, &site.MapEmbedURL,
			&site.FacebookURL, &site.InstagramURL, &site.WhatsAppURL,
			&site.TwitterURL, &site.TikTokURL, &site.LinkedInURL, &site.YouTubeURL,
			&site.UmamiWebsiteID, &site.LeadEmail, &site.Status, &site.CreatedAt, &site.PublishedAt,
			&site.Plan, &site.PaymentStatus, &site.StripeSessionID, &site.StripeSubscriptionID, &site.PaidAt,
			&site.CustomDomain, &site.Notes, &site.AnalyticsFrequency, &site.AnalyticsLastSent)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return site, err
}

func (s *Store) ListSites() ([]*models.Site, error) {
	rows, err := s.db.Query(`
		SELECT id, slug, business_name, template, tagline, about, services,
		       certifications, location, cta_text, testimonials, logo_url, gallery,
		       phone, email, address, hours, map_url, map_embed_url,
		       facebook_url, instagram_url, whatsapp_url, twitter_url, tiktok_url, linkedin_url, youtube_url,
		       umami_website_id, lead_email, status, created_at, published_at,
		       plan, payment_status, stripe_session_id, stripe_subscription_id, paid_at,
		       COALESCE(custom_domain, ''), notes, analytics_frequency, analytics_last_sent
		FROM sites ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []*models.Site
	for rows.Next() {
		site := &models.Site{}
		if err := rows.Scan(&site.ID, &site.Slug, &site.BusinessName, &site.Template,
			&site.Tagline, &site.About, &site.Services,
			&site.Certifications, &site.Location, &site.CTAText,
			&site.Testimonials, &site.LogoURL, &site.Gallery,
			&site.Phone, &site.Email, &site.Address, &site.Hours, &site.MapURL, &site.MapEmbedURL,
			&site.FacebookURL, &site.InstagramURL, &site.WhatsAppURL,
			&site.TwitterURL, &site.TikTokURL, &site.LinkedInURL, &site.YouTubeURL,
			&site.UmamiWebsiteID, &site.LeadEmail, &site.Status, &site.CreatedAt, &site.PublishedAt,
			&site.Plan, &site.PaymentStatus, &site.StripeSessionID, &site.StripeSubscriptionID, &site.PaidAt,
			&site.CustomDomain, &site.Notes, &site.AnalyticsFrequency, &site.AnalyticsLastSent); err != nil {
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

// --- Leads ---

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

func (s *Store) GetSiteByStripeSessionID(sessionID string) (*models.Site, error) {
	site := &models.Site{}
	err := s.db.QueryRow(`
		SELECT id, slug, business_name, template, tagline, about, services,
		       certifications, location, cta_text, testimonials, logo_url, gallery,
		       phone, email, address, hours, map_url, map_embed_url,
		       facebook_url, instagram_url, whatsapp_url, twitter_url, tiktok_url, linkedin_url, youtube_url,
		       umami_website_id, lead_email, status, created_at, published_at,
		       plan, payment_status, stripe_session_id, stripe_subscription_id, paid_at,
		       COALESCE(custom_domain, ''), notes, analytics_frequency, analytics_last_sent
		FROM sites WHERE stripe_session_id = $1`, sessionID).
		Scan(&site.ID, &site.Slug, &site.BusinessName, &site.Template,
			&site.Tagline, &site.About, &site.Services,
			&site.Certifications, &site.Location, &site.CTAText,
			&site.Testimonials, &site.LogoURL, &site.Gallery,
			&site.Phone, &site.Email, &site.Address, &site.Hours, &site.MapURL, &site.MapEmbedURL,
			&site.FacebookURL, &site.InstagramURL, &site.WhatsAppURL,
			&site.TwitterURL, &site.TikTokURL, &site.LinkedInURL, &site.YouTubeURL,
			&site.UmamiWebsiteID, &site.LeadEmail, &site.Status, &site.CreatedAt, &site.PublishedAt,
			&site.Plan, &site.PaymentStatus, &site.StripeSessionID, &site.StripeSubscriptionID, &site.PaidAt,
			&site.CustomDomain, &site.Notes, &site.AnalyticsFrequency, &site.AnalyticsLastSent)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return site, err
}

func (s *Store) GetSiteByStripeSubscriptionID(subID string) (*models.Site, error) {
	site := &models.Site{}
	err := s.db.QueryRow(`
		SELECT id, slug, business_name, template, tagline, about, services,
		       certifications, location, cta_text, testimonials, logo_url, gallery,
		       phone, email, address, hours, map_url, map_embed_url,
		       facebook_url, instagram_url, whatsapp_url, twitter_url, tiktok_url, linkedin_url, youtube_url,
		       umami_website_id, lead_email, status, created_at, published_at,
		       plan, payment_status, stripe_session_id, stripe_subscription_id, paid_at,
		       COALESCE(custom_domain, ''), notes, analytics_frequency, analytics_last_sent
		FROM sites WHERE stripe_subscription_id = $1`, subID).
		Scan(&site.ID, &site.Slug, &site.BusinessName, &site.Template,
			&site.Tagline, &site.About, &site.Services,
			&site.Certifications, &site.Location, &site.CTAText,
			&site.Testimonials, &site.LogoURL, &site.Gallery,
			&site.Phone, &site.Email, &site.Address, &site.Hours, &site.MapURL, &site.MapEmbedURL,
			&site.FacebookURL, &site.InstagramURL, &site.WhatsAppURL,
			&site.TwitterURL, &site.TikTokURL, &site.LinkedInURL, &site.YouTubeURL,
			&site.UmamiWebsiteID, &site.LeadEmail, &site.Status, &site.CreatedAt, &site.PublishedAt,
			&site.Plan, &site.PaymentStatus, &site.StripeSessionID, &site.StripeSubscriptionID, &site.PaidAt,
			&site.CustomDomain, &site.Notes, &site.AnalyticsFrequency, &site.AnalyticsLastSent)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return site, err
}

func (s *Store) SetSitePending(id int, plan, sessionID string) error {
	_, err := s.db.Exec(`UPDATE sites SET payment_status='pending', plan=$1, stripe_session_id=$2 WHERE id=$3`, plan, sessionID, id)
	return err
}

func (s *Store) SetSitePaid(sessionID, subscriptionID string) error {
	now := time.Now().UTC()
	_, err := s.db.Exec(`UPDATE sites SET payment_status='paid', paid_at=$1, stripe_subscription_id=$2 WHERE stripe_session_id=$3`, now, subscriptionID, sessionID)
	return err
}

func (s *Store) SetSiteCancelled(subscriptionID string) error {
	_, err := s.db.Exec(`UPDATE sites SET payment_status='cancelled' WHERE stripe_subscription_id=$1`, subscriptionID)
	return err
}

func (s *Store) GetSiteByCustomDomain(domain string) (*models.Site, error) {
	site := &models.Site{}
	err := s.db.QueryRow(`
		SELECT id, slug, business_name, template, tagline, about, services,
		       certifications, location, cta_text, testimonials, logo_url, gallery,
		       phone, email, address, hours, map_url, map_embed_url,
		       facebook_url, instagram_url, whatsapp_url, twitter_url, tiktok_url, linkedin_url, youtube_url,
		       umami_website_id, lead_email, status, created_at, published_at,
		       plan, payment_status, stripe_session_id, stripe_subscription_id, paid_at,
		       COALESCE(custom_domain, ''), notes, analytics_frequency, analytics_last_sent
		FROM sites WHERE custom_domain = $1`, domain).
		Scan(&site.ID, &site.Slug, &site.BusinessName, &site.Template,
			&site.Tagline, &site.About, &site.Services,
			&site.Certifications, &site.Location, &site.CTAText,
			&site.Testimonials, &site.LogoURL, &site.Gallery,
			&site.Phone, &site.Email, &site.Address, &site.Hours, &site.MapURL, &site.MapEmbedURL,
			&site.FacebookURL, &site.InstagramURL, &site.WhatsAppURL,
			&site.TwitterURL, &site.TikTokURL, &site.LinkedInURL, &site.YouTubeURL,
			&site.UmamiWebsiteID, &site.LeadEmail, &site.Status, &site.CreatedAt, &site.PublishedAt,
			&site.Plan, &site.PaymentStatus, &site.StripeSessionID, &site.StripeSubscriptionID, &site.PaidAt,
			&site.CustomDomain, &site.Notes, &site.AnalyticsFrequency, &site.AnalyticsLastSent)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return site, err
}

func (s *Store) SetCustomDomain(id int, domain string) error {
	var val *string
	if domain != "" {
		val = &domain
	}
	_, err := s.db.Exec(`UPDATE sites SET custom_domain = $1 WHERE id = $2`, val, id)
	return err
}

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

func (s *Store) UpdateAnalyticsLastSent(id int) error {
	_, err := s.db.Exec(`UPDATE sites SET analytics_last_sent = NOW() WHERE id = $1`, id)
	return err
}

func (s *Store) UpdateAnalyticsFrequency(id int, freq string) error {
	_, err := s.db.Exec(`UPDATE sites SET analytics_frequency = $1 WHERE id = $2`, freq, id)
	return err
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
