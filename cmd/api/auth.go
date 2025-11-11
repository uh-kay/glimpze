package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uh-kay/glimpze/auth"
	"github.com/uh-kay/glimpze/store"
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

	auth.SetAuthCookies(w, token)
	app.jsonResponse(w, http.StatusOK, envelope{
		Message: "success",
		Data:    user,
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

	err := app.store.Users.Create(r.Context(), user)
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

	app.jsonResponse(w, http.StatusCreated, envelope{
		Message: "user registered",
		Data:    user,
	})
}

func (app *application) profile(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	app.jsonResponse(w, http.StatusOK, envelope{
		Message: "success",
		Data:    user,
	})
}

func getUserFromContext(r *http.Request) *store.User {
	return r.Context().Value(userCtx).(*store.User)
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
		Message: "success",
		Data:    nil,
	})
}
