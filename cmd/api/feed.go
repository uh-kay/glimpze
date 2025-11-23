package main

import (
	"net/http"
	"strconv"

	"newsdrop.org/store"
)

func (app *application) userFeed(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)

	pageStr := r.URL.Query().Get("page")
	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var posts []*store.PostWithMetadata
	if user != nil {
		posts, err = app.store.Posts.GetUserFeed(r.Context(), user.ID, 20, (page-1)*20)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
	} else {
		posts, err = app.store.Posts.GetPublicFeed(r.Context(), 20, (page-1)*20)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
	}

	app.jsonResponse(w, http.StatusOK, envelope{
		"message": "success",
		"posts":   posts,
	})
}
