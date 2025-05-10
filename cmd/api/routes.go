package main

import (
	"net/http"
	"net/http/pprof"
)

type handlerFunc func(w http.ResponseWriter, r *http.Request)

func validateAPIKey(app *Application, finalHandler handlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.requestHasInvalidAPIKey(r) {
			app.invalidAPIKeyResponse(w, r)
			return
		}
		finalHandler(w, r)
	})
}

func registerPprofHandlers(mux *http.ServeMux) { // nolint:unused
	// Register pprof handlers
	// import "net/http/pprof"
	// Tutorial: https://medium.com/@rahul.fiem/application-performance-optimization-how-to-effectively-analyze-and-optimize-pprof-cpu-profiles-95280b2f5bfb
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

func (app *Application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /api/where/agencies-with-coverage.json", validateAPIKey(app, app.agenciesWithCoverageHandler))
	mux.Handle("GET /api/where/agency/{id}", validateAPIKey(app, app.agencyHandler))
	mux.Handle("GET /api/where/current-time.json", validateAPIKey(app, app.currentTimeHandler))
	mux.Handle("GET /api/where/routes-for-agency/{id}", validateAPIKey(app, app.routesForAgencyHandler))

	return mux
}
