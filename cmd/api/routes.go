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

	// Register pprof handlers - https://medium.com/@rahul.fiem/application-performance-optimization-how-to-effectively-analyze-and-optimize-pprof-cpu-profiles-95280b2f5bfb
	// 	"net/http/pprof"
	// Register pprof handlers
	// tutorial: https://medium.com/@rahul.fiem/application-performance-optimization-how-to-effectively-analyze-and-optimize-pprof-cpu-profiles-95280b2f5bfb
	// import "net/http/pprof"
	//mux.HandleFunc("/debug/pprof/", pprof.Index)
	//mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	//mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	//mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	//mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return mux
}
