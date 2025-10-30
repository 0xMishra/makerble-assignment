package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	// Convert the notFoundResponse() helper to a http.Handler using the
	// http.HandlerFunc() adapter, and then set it as the custom error handler for 404
	// Not Found responses.
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	// Likewise, convert the methodNotAllowedResponse() helper to a http.Handler and set
	// it as the custom error handler for 405 Method Not Allowed responses.
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/register", app.registerHandler)

	router.HandlerFunc(http.MethodPost, "/v1/patients", app.authenticate("receptionist", "", app.addPatientHandler))
	router.HandlerFunc(http.MethodGet, "/v1/patients/:id", app.authenticate("receptionist", "doctor", app.getPatientHandler))
	router.HandlerFunc(http.MethodPut, "/v1/patients/:id", app.authenticate("receptionist", "doctor", app.updatePatientHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/patients/:id", app.authenticate("receptionist", "", app.deletePatientHandler))

	return app.recoverPanic(app.enableCORS(app.rateLimit(router)))
}
