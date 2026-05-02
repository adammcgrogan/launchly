package handlers

import (
	"net/http"
	"net/url"
)

const flashCookie = "_flash"

func setFlash(w http.ResponseWriter, msg string) {
	http.SetCookie(w, &http.Cookie{
		Name:     flashCookie,
		Value:    url.QueryEscape(msg),
		Path:     "/admin",
		MaxAge:   30,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func getFlash(w http.ResponseWriter, r *http.Request) string {
	c, err := r.Cookie(flashCookie)
	if err != nil {
		return ""
	}
	http.SetCookie(w, &http.Cookie{
		Name:   flashCookie,
		Path:   "/admin",
		MaxAge: -1,
	})
	msg, _ := url.QueryUnescape(c.Value)
	return msg
}
