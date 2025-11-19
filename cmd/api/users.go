package main

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
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
		err = app.store.UserLimits.Decrement(r.Context(), user.ID, "follow_limit")
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
		"message":  "user followed",
		"follower": follower,
	})
}

func (app *application) unfollowUser(w http.ResponseWriter, r *http.Request) {
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

	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		if err = s.Followers.Delete(r.Context(), followingID, user.ID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		if err = s.UserLimits.Increment(r.Context(), user.ID, "follow_limit"); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type ProfileForm struct {
	Biodata string `json:"biodata" validate:"required,min=1,max=255"`
}

func (app *application) createProfile(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	if err := r.ParseMultipartForm(maxFormSize); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	biodata := r.PostFormValue("biodata")
	if err := Validate.Struct(ProfileForm{Biodata: biodata}); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.validateFileUpload(fileHeader); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	defer file.Close()

	fileID := uuid.New()
	fileExt := filepath.Ext(fileHeader.Filename)
	filename := fmt.Sprintf("%s%s", fileID, fileExt)

	var userProfile *store.UserProfile
	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		userProfile, err = s.UserProfiles.Create(r.Context(), fileID, fileExt, user.ID, biodata)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.storage.SaveToR2(r.Context(), file, fileExt, filename); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusCreated, envelope{
		"message":      "user profile created",
		"user_profile": userProfile,
	})
}

func (app *application) updateProfile(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	if err := r.ParseMultipartForm(maxFormSize); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	biodata := r.PostFormValue("biodata")
	if err := Validate.Struct(ProfileForm{Biodata: biodata}); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil && err != http.ErrMissingFile {
		app.badRequestResponse(w, r, err)
		return
	}

	hasNewFile := err == nil
	if hasNewFile {
		defer file.Close()

		if err := app.validateFileUpload(fileHeader); err != nil {
			app.badRequestResponse(w, r, err)
			return
		}
	}

	oldUserProfile, err := app.store.UserProfiles.GetByUserID(r.Context(), user.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	var userProfile *store.UserProfile

	if hasNewFile {
		fileID := uuid.New()
		fileExt := filepath.Ext(fileHeader.Filename)
		newFilename := fmt.Sprintf("%s%s", fileID, fileExt)
		oldFilename := fmt.Sprintf("%s%s", oldUserProfile.FileID.String(), oldUserProfile.FileExtension)

		if err := app.storage.SaveToR2(r.Context(), file, fileExt, newFilename); err != nil {
			app.internalServerError(w, r, err)
			return
		}

		err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
			userProfile, err = s.UserProfiles.Update(r.Context(), fileID, fileExt, user.ID, biodata)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			_ = app.storage.DeleteFromR2(r.Context(), newFilename)
			app.internalServerError(w, r, err)
			return
		}

		if err = app.storage.DeleteFromR2(r.Context(), oldFilename); err != nil {
			app.internalServerError(w, r, err)
			return
		}
	} else {
		err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
			userProfile, err = s.UserProfiles.Update(r.Context(), oldUserProfile.FileID, oldUserProfile.FileExtension, user.ID, biodata)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			app.internalServerError(w, r, err)
		}
	}

	app.jsonResponse(w, http.StatusOK, envelope{
		"message":      "profile updated",
		"user_profile": userProfile,
	})
}
