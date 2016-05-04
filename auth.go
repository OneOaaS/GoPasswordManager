package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"goji.io"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
)

func PostLogin(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(rw, "could not parse form", http.StatusBadRequest)
		return
	}

	config := ConfigFromContext(ctx)

	id := r.FormValue("username")
	pass := r.FormValue("password")
	if id == "" || pass == "" {
		http.Error(rw, "missing parameter", http.StatusBadRequest)
		return
	}

	u, err := GetUser(ctx, id)
	if err != nil {
		http.Error(rw, "unknown user or bad password", http.StatusForbidden)
		return
	}

	if err := bcrypt.CompareHashAndPassword(u.Password, []byte(pass)); err != nil {
		http.Error(rw, "unknown user or bad password", http.StatusForbidden)
		return
	}

	s := Session{
		UserID: u.ID,
		Time:   time.Now(),
	}
	s.Sign(config.CookieSecret)

	b, err := json.Marshal(s)
	if err != nil {
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(rw, &http.Cookie{
		Name:  config.CookieName,
		Value: base64.URLEncoding.EncodeToString(b),
	})

	// http.Redirect(
}

// Auth resumes a session.
func Auth(next goji.Handler) goji.Handler {
	return goji.HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
		config := ConfigFromContext(ctx)
		var s Session

		// verify session cookie
		if cookie, err := r.Cookie(config.CookieName); err != nil {
			http.Error(rw, "missing auth cookie", http.StatusBadRequest)
			return
		} else if b, err := base64.URLEncoding.DecodeString(cookie.Value); err != nil {
			http.Error(rw, "invalid auth cookie", http.StatusBadRequest)
			return
		} else if err := json.Unmarshal(b, &s); err != nil {
			http.Error(rw, "invalid auth cookie", http.StatusBadRequest)
			return
		} else if !s.Verify(config.CookieSecret) {
			http.Error(rw, "invalid auth cookie", http.StatusBadRequest)
			return
		}
		// TODO(tolar2): enforce session timestamp?

		u, err := GetUser(ctx, s.UserID)
		if err != nil {
			http.Error(rw, "unknown user", http.StatusBadRequest)
			return
		}

		ctx = ContextWithSession(ctx, s)
		ctx = ContextWithUser(ctx, u)

		next.ServeHTTPC(ctx, rw, r)
	})
}
