package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/uh-kay/glimpze/auth"
)

type contextKey string

const userCtx contextKey = "user"

func bearerFromHeader(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if after, ok := strings.CutPrefix(h, "Bearer "); ok {
		return after
	}
	return ""
}

func (app *application) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenStr string

		cookie, err := r.Cookie("access_token")
		if err == nil && cookie != nil {
			tokenStr = cookie.Value
		}

		if tokenStr == "" {
			tokenStr = bearerFromHeader(r)
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

		userIDStr, err := app.cache.Sessions.GetUser(r.Context(), "access:"+claims.ID)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		user, err := app.store.Users.GetByID(r.Context(), int64(userID))
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		ctx := context.WithValue(r.Context(), userCtx, user)
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
