package main

import (
	"errors"
	"net/http"
	"strconv"

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
		Message: "success",
		Data:    user,
	})
}

func (app *application) followUser(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	followingIDStr := r.PathValue("userID")
	followingID, err := strconv.ParseInt(followingIDStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if user.ID == followingID {
		app.badRequestResponse(w, r, errors.New("can't follow yourself"))
		return
	}

	var follower *store.Follower
	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		follower, err = s.Followers.Create(r.Context(), followingID, user.ID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		err = app.store.UserLimits.Reduce(r.Context(), user.ID, "follow_limit")
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusCreated, envelope{
		Message: "user followed",
		Data:    follower,
	})
}

func (app *application) unfollowUser(w http.ResponseWriter, r *http.Request) {}
