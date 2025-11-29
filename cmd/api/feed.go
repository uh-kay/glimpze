package main

import (
	"fmt"
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

	for _, post := range posts {
		links := make([]string, 0, len(post.Post.FileIDs))
		for j := range len(post.Post.FileIDs) {
			link, err := app.storage.GetFromR2(r.Context(), fmt.Sprintf("%s%s", post.Post.FileIDs[j].String(), post.Post.FileExtensions[j]))
			if err != nil {
				app.internalServerError(w, r, err)
				return
			}
			links = append(links, link)
		}

		post.ImageLinks = links
	}

	app.jsonResponse(w, http.StatusOK, envelope{
		"message": "success",
		"posts":   posts,
	})
}
