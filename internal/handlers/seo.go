package handlers

import (
	"fmt"
	"net/http"
)

func (h *Handler) RobotsTxt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "User-agent: *\nAllow: /\n\nDisallow: /admin\nDisallow: /sites/\n\nSitemap: https://%s/sitemap.xml\n", h.domain)
}

func (h *Handler) Privacy(w http.ResponseWriter, r *http.Request) {
	h.render(w, "privacy", nil)
}

func (h *Handler) Terms(w http.ResponseWriter, r *http.Request) {
	h.render(w, "terms", nil)
}

func (h *Handler) Sitemap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://%s/</loc>
    <changefreq>weekly</changefreq>
    <priority>1.0</priority>
  </url>
  <url>
    <loc>https://%s/get-started</loc>
    <changefreq>monthly</changefreq>
    <priority>0.8</priority>
  </url>
  <url>
    <loc>https://%s/privacy</loc>
    <changefreq>yearly</changefreq>
    <priority>0.2</priority>
  </url>
  <url>
    <loc>https://%s/terms</loc>
    <changefreq>yearly</changefreq>
    <priority>0.2</priority>
  </url>
`, h.domain, h.domain, h.domain, h.domain)

	if sites, err := h.store.ListLiveSites(); err == nil {
		for _, s := range sites {
			loc := "https://" + s.Slug + "." + h.domain
			if s.CustomDomain != "" {
				loc = "https://" + s.CustomDomain
			}
			lastmod := ""
			if s.PublishedAt != nil {
				lastmod = fmt.Sprintf("\n    <lastmod>%s</lastmod>", s.PublishedAt.Format("2006-01-02"))
			}
			fmt.Fprintf(w, "  <url>\n    <loc>%s</loc>%s\n    <changefreq>monthly</changefreq>\n    <priority>0.6</priority>\n  </url>\n", loc, lastmod)
		}
	}

	fmt.Fprint(w, "</urlset>\n")
}
