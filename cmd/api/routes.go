package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	// Initialize a new httprouter router instance.
	router := httprouter.New()

	// Register the relevant methods, URL patterns and handler functions for our
	// endpoints using the HandlerFunc() method. Note that http.MethodGet and
	// http.MethodPost are constants which equate to the strings "GET" and "POST"
	// respectively.
	router.HandlerFunc(http.MethodGet, "/api/where/agencies-with-coverage.json", app.agenciesWithCoverageHandler)
	router.HandlerFunc(http.MethodGet, "/api/where/agency/:id.json", app.agencyHandler)
	router.HandlerFunc(http.MethodGet, "/api/where/current-time.json", app.currentTimeHandler)

	return router
}
