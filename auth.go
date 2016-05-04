package main

import (
	"encoding/json"
	"net/http"

	"goji.io"

	"golang.org/x/net/context"
)

// Auth authenticates a request
func Auth(next goji.Handler) goji.Handler {
	return goji.HandlerFunc(func(ctx context.Context, rw http.ResponseWriter, r *http.Request) {
		config := ConfigFromContext(ctx)
		var s Session

		// verify session cookie
		if cookie, err := r.Cookie(config.CookieName); err != nil {
			http.Error(rw, "missing auth cookie", http.StatusBadRequest)
			return
		} else if err := json.Unmarshal([]byte(cookie.Value), &s); err != nil {
			http.Error(rw, "invalid auth cookie", http.StatusBadRequest)
			return
		} else if !s.Verify(config.CookieSecret) {
			http.Error(rw, "invalid auth cookie", http.StatusBadRequest)
			return
		}
		// TODO: enforce session timestamp?

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
