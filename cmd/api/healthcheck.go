package main

import "net/http"

func (app *application) healthcheck(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "available",
		"version": version,
	}

	writeJSON(w, http.StatusOK, data)
}
