package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"newsdrop.org/auth"
	"newsdrop.org/store"
)

type LoginPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=6,max=72"`
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	var payload LoginPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := user.Password.Compare(payload.Password); err != nil {
		app.unauthorizedErrorResponse(w, r, err)
		return
	}

	token, err := auth.IssueTokens(strconv.FormatInt(user.ID, 10))
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := auth.Persist(r.Context(), &app.cache, token); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// auth.SetAuthCookies(w, token)
	app.jsonResponse(w, http.StatusOK, envelope{
		"message":       "success",
		"user":          user,
		"access_token":  token.Access,
		"refresh_token": token.Refresh,
	})
}

type RegisterPayload struct {
	Name        string `json:"name" binding:"required,min=2,max=30" example:"test"`
	DisplayName string `json:"display_name" binding:"required,min=2,max=30" example:"test"`
	Email       string `json:"email" binding:"required,email,max=255" example:"test@test.com"`
	Password    string `json:"password" binding:"required,min=6,max=72" example:"password"`
}

func (app *application) register(w http.ResponseWriter, r *http.Request) {
	var payload RegisterPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		Name:        payload.Name,
		DisplayName: payload.DisplayName,
		Email:       payload.Email,
		Role:        *app.defaultRole,
	}

	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	var err error
	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		err = app.store.Users.Create(r.Context(), user)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_email_key":
				app.badRequestResponse(w, r, ErrDuplicateEmail)
			case "users_name_key":
				app.badRequestResponse(w, r, ErrDuplicateName)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	token, err := app.store.Tokens.New(user.ID, 3*24*time.Hour, store.ScopeActivation)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	isProd := app.config.env == "prod"

	app.background(func() {
		data := map[string]any{
			"Username":      user.Name,
			"ActivationURL": fmt.Sprintf("http://localhost:5173/activate?token=%s", token.Plaintext),
		}

		err = app.mailer.SendAPI("user_welcome.tmpl", user.Name, user.Email, data, !isProd)
		if err != nil {
			app.internalServerError(w, r, err)
		}
	})

	app.jsonResponse(w, http.StatusCreated, envelope{
		"message": "user registered",
		"user":    user,
	})
}

func getUserFromContext(r *http.Request) *store.User {
	user, ok := r.Context().Value(userCtx).(*store.User)
	if !ok {
		return nil
	}
	return user
}

func (app *application) refreshToken(w http.ResponseWriter, r *http.Request) {
	ref, err := auth.MustCookie(r, "refresh_token")
	if err != nil {
		app.unauthorizedErrorResponse(w, r, err)
		return
	}

	claims, err := auth.ParseRefresh(ref)
	if err != nil {
		app.unauthorizedErrorResponse(w, r, err)
		return
	}
	if _, err := app.cache.Sessions.GetUser(r.Context(), "refresh:"+claims.ID); err != nil {
		app.unauthorizedErrorResponse(w, r, err)
		return
	}
	_ = app.cache.Sessions.Delete(r.Context(), "refresh:"+claims.ID)

	toks, err := auth.IssueTokens(claims.Subject)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := auth.Persist(r.Context(), &app.cache, toks); err != nil {
		app.internalServerError(w, r, err)
		return
	}
	auth.SetAuthCookies(w, toks)
	app.jsonResponse(w, http.StatusCreated, envelope{
		"message": "success",
	})
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	accessCookie, err := r.Cookie("access_token")
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	acc := accessCookie.Value

	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	ref := refreshCookie.Value

	if acc != "" {
		if claims, err := auth.ParseAccess(acc); err == nil {
			_ = app.cache.Sessions.Delete(r.Context(), "access"+claims.ID)
		}
	}

	if ref != "" {
		if claims, err := auth.ParseRefresh(ref); err == nil {
			_ = app.cache.Sessions.Delete(r.Context(), "refresh"+claims.ID)
		}
	}
	auth.ClearAuthCookies(w)
	app.jsonResponse(w, http.StatusOK, envelope{
		"message": "success",
	})
}

type ActivatePayload struct {
	Token string `json:"token" validate:"required"`
}

func (app *application) activateUser(w http.ResponseWriter, r *http.Request) {
	var payload ActivatePayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.store.Users.GetByToken(store.ScopeActivation, payload.Token)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	user.Activated = true

	err = app.store.Users.Update(user)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	err = app.store.Tokens.Delete(store.ScopeActivation, user.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, envelope{
		"user": user,
	})
}
