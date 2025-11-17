package main

import (
	"context"
	"errors"
	"fmt"
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

	fmt.Printf("User role: %s (level %d), Required role: %s (level %d)\n",
		user.Role.Name, user.Role.Level, rolename, role.Level)

	return user.Role.Level >= role.Level, nil
}

type LimitType string

const (
	CreatePostLimit LimitType = "create_post"
	CommentLimit    LimitType = "comment"
	LikeLimit       LimitType = "like"
	FollowLimit     LimitType = "follow"
)

func (app *application) checkLimit(user *store.User, limitType LimitType) bool {
	switch limitType {
	case CreatePostLimit:
		return user.UserLimit.CreatePostLimit > 0
	case CommentLimit:
		return user.UserLimit.CommentLimit > 0
	case LikeLimit:
		return user.UserLimit.LikeLimit > 0
	case FollowLimit:
		return user.UserLimit.FollowLimit > 0
	default:
		return true
	}
}

func (app *application) checkResourceAccessWithLimit(requiredRole string, limitType LimitType, next http.HandlerFunc) http.HandlerFunc {
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

		limitAllowed := app.checkLimit(user, limitType)
		if !limitAllowed {
			app.forbiddenResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) optionalAuthMiddleware(next http.Handler) http.Handler {
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
			next.ServeHTTP(w, r)
			return
		}

		claims, err := auth.ParseAccess(tokenStr)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		userIDStr, err := app.cache.Sessions.GetUser(r.Context(), "access:"+claims.ID)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		user, err := app.store.Users.GetByID(r.Context(), int64(userID))
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
