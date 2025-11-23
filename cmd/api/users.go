package main

import (
	"net/http"
	"strconv"

	"newsdrop.org/store"
)

type UpdateUserRolePayload struct {
	RoleName string `json:"role_name" validate:"required,min=1,max=255"`
}

func (app *application) updateUserRole(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("userName")

	var payload UpdateUserRolePayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	role, err := app.store.Roles.GetByName(r.Context(), payload.RoleName)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	var user *store.User
	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		user, err = app.store.Users.UpdateRole(r.Context(), name, role)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, envelope{
		"message": "role updated",
		"user":    user,
	})
}

func (app *application) profile(w http.ResponseWriter, r *http.Request) {
	var user *store.User
	userIDStr := r.PathValue("userID")
	if userIDStr == "" {
		user = getUserFromContext(r)
	} else {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}

		user, err = app.store.Users.GetByID(r.Context(), userID)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
	}

	app.jsonResponse(w, http.StatusOK, envelope{
		"message": "success",
		"user":    user,
	})
}
