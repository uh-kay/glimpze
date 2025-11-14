package main

import (
	"net/http"

	"github.com/uh-kay/glimpze/store"
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
		Message: "role updated",
		Data:    user,
	})
}
