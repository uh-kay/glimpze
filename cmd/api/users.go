package main

import "net/http"

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

	tx, err := app.db.Begin(r.Context())
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	defer tx.Rollback(r.Context())

	role, err := app.store.Roles.GetByName(r.Context(), payload.RoleName)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	user, err := app.store.Users.UpdateRole(r.Context(), tx, name, role)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, envelope{
		Message: "role updated",
		Data:    user,
	})
}
