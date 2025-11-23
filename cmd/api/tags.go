package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgconn"
	"newsdrop.org/store"
)

type TagPayload struct {
	Name string `json:"name" validate:"required,min=1,max=50"`
}

func (app *application) createTag(w http.ResponseWriter, r *http.Request) {
	var payload TagPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var tag *store.Tag
	var err error
	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		tag, err = app.store.Tags.Create(r.Context(), payload.Name)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "tags_name_key":
				app.conflictError(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusCreated, envelope{
		"message": "tag created",
		"tag":     tag,
	})
}

func (app *application) getTag(w http.ResponseWriter, r *http.Request) {
	tagIDStr := r.PathValue("tagID")
	tagID, err := strconv.ParseInt(tagIDStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	tag, err := app.store.Tags.GetByID(r.Context(), tagID)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	app.jsonResponse(w, http.StatusOK, envelope{
		"message": "success",
		"tag":     tag,
	})
}

func (app *application) deleteTag(w http.ResponseWriter, r *http.Request) {
	tagIDStr := r.PathValue("tagID")
	tagID, err := strconv.ParseInt(tagIDStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		err = app.store.Tags.Delete(r.Context(), tagID)
		if err != nil {
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

func (app *application) addTag(w http.ResponseWriter, r *http.Request) {
	var payload TagPayload

	postIDStr := r.PathValue("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	tag, err := app.store.Tags.GetByName(r.Context(), payload.Name)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	var postTag *store.PostTag
	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		postTag, err = app.store.PostTags.Create(r.Context(), postID, tag.ID, tag.Name)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "post_tags_post_id_tag_id_key", "post_tags_pkey":
				app.conflictError(w, r, err)
				return
			default:
				app.internalServerError(w, r, err)
				return
			}
		}

		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	app.jsonResponse(w, http.StatusCreated, envelope{
		"message":  "tag added",
		"post_tag": postTag,
	})
}

func (app *application) removeTag(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.PathValue("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	tagIDStr := r.PathValue("tagID")
	tagID, err := strconv.ParseInt(tagIDStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		err := app.store.PostTags.Delete(r.Context(), postID, tagID)
		if err != nil {
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

func (app *application) listTag(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.PathValue("postID")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	postTags, err := app.store.PostTags.List(r.Context(), postID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusOK, envelope{
		"message":   "success",
		"post_tags": postTags,
	})
}
