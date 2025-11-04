package main

import (
	"net/http"
)

type PostPayload struct {
	Title   string `json:"title" validate:"required,min=1,max=255"`
	Content string `json:"content" validate:"required,min=1,max=2048"`
}

func (app *application) createPost(w http.ResponseWriter, r *http.Request) {
	var payload PostPayload

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	post, err := app.store.Posts.Create(r.Context(), payload.Title, payload.Content)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusCreated, envelope{
		Message: "post created",
		Data:    post,
	})
}
