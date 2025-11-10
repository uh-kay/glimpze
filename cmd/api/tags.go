package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uh-kay/glimpze/store"
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

	tx, err := app.db.Begin(r.Context())
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	defer tx.Rollback(r.Context())

	tag, err := app.store.Tags.Create(r.Context(), tx, payload.Name)
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

	if err := tx.Commit(r.Context()); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusCreated, envelope{
		Message: "tag created",
		Data:    tag,
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
		Message: "success",
		Data:    tag,
	})
}

func (app *application) deleteTag(w http.ResponseWriter, r *http.Request) {
	tagIDStr := r.PathValue("tagID")
	tagID, err := strconv.ParseInt(tagIDStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	tx, err := app.db.Begin(r.Context())
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	defer tx.Rollback(r.Context())

	if err := app.store.Tags.Delete(r.Context(), tx, tagID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
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

	tx, err := app.db.Begin(r.Context())
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	defer tx.Rollback(r.Context())

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

	postTag, err := app.store.PostTags.Create(r.Context(), tx, postID, tag.ID, tag.Name)
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

	if err := tx.Commit(r.Context()); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusCreated, envelope{
		Message: "tag added",
		Data:    postTag,
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

	tx, err := app.db.Begin(r.Context())
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	defer tx.Rollback(r.Context())

	if err := app.store.PostTags.Delete(r.Context(), tx, postID, tagID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
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
		Message: "success",
		Data:    postTags,
	})
}
