package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

const sessionCookie = "_session"

func (h *Handler) deriveToken(ctx string) string {
	mac := hmac.New(sha256.New, []byte(h.adminPass))
	mac.Write([]byte(ctx))
	return hex.EncodeToString(mac.Sum(nil))
}

func (h *Handler) csrfToken() string {
	return h.deriveToken("csrf-v1")[:32]
}

func (h *Handler) setSession(w http.ResponseWriter) {
	secure := !strings.Contains(h.domain, "localhost") && !strings.HasPrefix(h.domain, "127.")
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    h.deriveToken("session-v1"),
		Path:     "/admin",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   sessionCookie,
		Path:   "/admin",
		MaxAge: -1,
	})
}

func (h *Handler) isAuthenticated(r *http.Request) bool {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return false
	}
	return hmac.Equal([]byte(c.Value), []byte(h.deriveToken("session-v1")))
}
