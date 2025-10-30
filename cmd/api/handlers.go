package main

import (
	"net/http"
)

func (app *application) registerHandler(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get("Role")

	switch role {
	case "doctor":
		app.registerDoctorHandler(w, r)

	case "receptionist":
		app.registerReceptionistHandler(w, r)

	default:
		app.badRequestResponse(w, r, http.ErrAbortHandler)
	}
}
