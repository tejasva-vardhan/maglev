package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/where/agencies-with-coverage.json", app.agenciesWithCoverageHandler)
	mux.HandleFunc("GET /api/where/agency/{id}", app.agencyHandler)
	mux.HandleFunc("GET /api/where/current-time.json", app.currentTimeHandler)
	mux.HandleFunc("GET /api/where/routes-for-agency/{id}", app.routesForAgencyHandler)

	return mux
}
