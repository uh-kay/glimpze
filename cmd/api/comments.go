package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/uh-kay/glimpze/store"
)

type CommentPayload struct {
	Content string `json:"content" validate:"required,min=1,max=2048"`
}

func (app *application) createComment(w http.ResponseWriter, r *http.Request) {
	var payload CommentPayload

	user := getUserFromContext(r)
	post := getPostFromContext(r)

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var comment *store.Comment
	var err error
	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		comment, err = app.store.Comments.Create(r.Context(), payload.Content, user.ID, post.ID)
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
		Message: "comment created",
		Data:    comment,
	})
}

func (app *application) getComment(w http.ResponseWriter, r *http.Request) {
	commentIDStr := r.PathValue("commentID")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	comment, err := app.store.Comments.GetByID(r.Context(), commentID)
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
		Data:    comment,
	})
}

func (app *application) updateComment(w http.ResponseWriter, r *http.Request) {
	var payload CommentPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	commentIDStr := r.PathValue("commentID")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var comment *store.Comment
	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		comment, err = app.store.Comments.Update(r.Context(), payload.Content, commentID)
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
		Message: "comment updated",
		Data:    comment,
	})
}

func (app *application) deleteComment(w http.ResponseWriter, r *http.Request) {
	commentIDStr := r.PathValue("commentID")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.store.WithTx(r.Context(), func(s *store.Storage) error {
		err := app.store.Comments.Delete(r.Context(), commentID)
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
