package main

import (
	"encoding/json"
	"net/http"
	"time"

	"goji.io"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
)

func PostLogin(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
	config := ConfigFromContext(ctx)

	var params struct {
		Username string
		Password string
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(rw, "invalid JSON", http.StatusBadRequest)
		return
	}

	id := params.Username
	pass := params.Password
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

	if value, err := SecureCookieFromContext(ctx).Encode(config.CookieName, s); err != nil {
		http.Error(rw, "internal server error", http.StatusInternalServerError)
		return
	} else {
		http.SetCookie(rw, &http.Cookie{
			Name:  config.CookieName,
			Value: value,
		})
	}
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
		} else if err := SecureCookieFromContext(ctx).Decode(config.CookieName, cookie.Value, &s); err != nil {
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
