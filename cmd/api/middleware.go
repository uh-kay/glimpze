package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/uh-kay/glimpze/auth"
	"github.com/uh-kay/glimpze/store"
)

type contextKey string

const userCtx contextKey = "user"
const postCtx contextKey = "post"

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

func (app *application) checkPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)
		post := getPostFromContext(r)

		if post.UserID == user.ID {
			next.ServeHTTP(w, r)
			return
		}

		allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if !allowed {
			app.forbiddenResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) checkResourceAccess(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)

		allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if !allowed {
			app.forbiddenResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) checkRolePrecedence(ctx context.Context, user *store.User, rolename string) (bool, error) {
	role, err := app.store.Roles.GetByName(ctx, rolename)
	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil
}
