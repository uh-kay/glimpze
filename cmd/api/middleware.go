package main

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/uh-kay/glimpze/auth"
	"github.com/uh-kay/glimpze/store/cache"
)

type contextKey string

const userIDCtx contextKey = "userID"

func bearerFromHeader(w http.ResponseWriter) string {
	h := w.Header().Get("Authorization")
	if after, ok := strings.CutPrefix(h, "Bearer "); ok {
		return after
	}
	return ""
}

func (app *application) AuthMiddleware(v *cache.Storage, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, _ := r.Cookie("access_token")
		tokenStr := cookie.Value
		if tokenStr == "" {
			tokenStr = bearerFromHeader(w)
		}
		if tokenStr == "" {
			app.unauthorizedErrorResponse(w, r, errors.New("missing token"))
			return
		}

		claims, err := auth.ParseAccess(tokenStr)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		if _, err := v.Sessions.GetUser(r.Context(), "access:"+claims.ID); err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		ctx := context.WithValue(r.Context(), userIDCtx, claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func MustCookie(w http.ResponseWriter, r *http.Request, name string) (string, error) {
	val, err := r.Cookie(name)
	valStr := val.Value
	if err != nil || valStr == "" {
		return "", errors.New("missing cookie: " + name)
	}
	return valStr, nil
}
