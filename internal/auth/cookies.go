package auth

import (
	"net/http"
	"time"
)

type CookieCfg struct {
	Domain string
}

func setCookie(w http.ResponseWriter, name, val string, cfg CookieCfg, ttl time.Duration, httpOnly bool) {
	ck := &http.Cookie{
		Name:     name,
		Value:    val,
		Domain:   cfg.Domain,
		Path:     "/",
		Secure:   true,
		HttpOnly: httpOnly,
		SameSite: http.SameSiteNoneMode,
	}
	if ttl > 0 {
		ck.Expires = time.Now().Add(ttl)
	}
	http.SetCookie(w, ck)
}

func delCookie(w http.ResponseWriter, name string, cfg CookieCfg) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Domain:   cfg.Domain,
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})
}
