package restapi

import (
	"net/http"
	"net/http/pprof"
)

type handlerFunc func(w http.ResponseWriter, r *http.Request)

func validateAPIKey(api *RestAPI, finalHandler handlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if api.RequestHasInvalidAPIKey(r) {
			api.invalidAPIKeyResponse(w, r)
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

func (api *RestAPI) SetRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/where/agencies-with-coverage.json", validateAPIKey(api, api.agenciesWithCoverageHandler))
	mux.Handle("GET /api/where/agency/{id}", validateAPIKey(api, api.agencyHandler))
	mux.Handle("GET /api/where/current-time.json", validateAPIKey(api, api.currentTimeHandler))
	mux.Handle("GET /api/where/routes-for-agency/{id}", validateAPIKey(api, api.routesForAgencyHandler))
	mux.Handle("GET /api/where/vehicles-for-agency/{id}", validateAPIKey(api, api.vehiclesForAgencyHandler))
	mux.Handle("GET /api/where/stops-for-location.json", validateAPIKey(api, api.stopsForLocationHandler))
	mux.Handle("GET /api/where/stop-ids-for-agency/{id}", validateAPIKey(api, api.stopIDsForAgencyHandler))
	mux.Handle("GET /api/where/report-problem-with-trip/{id}", validateAPIKey(api, api.reportProblemWithTripHandler))
	mux.Handle("GET /api/where/report-problem-with-stop/{id}", validateAPIKey(api, api.reportProblemWithStopHandler))
	mux.Handle("GET /api/where/trip/{id}", validateAPIKey(api, api.tripHandler))
	mux.Handle("GET /api/where/route-ids-for-agency/{id}", validateAPIKey(api, api.routeIDsForAgencyHandler))
	mux.Handle("GET /api/where/stop/{id}", validateAPIKey(api, api.stopHandler))
	mux.Handle("GET /api/where/shape/{id}", validateAPIKey(api, api.shapes))
	mux.Handle("GET /api/where/routes-for-location.json", validateAPIKey(api, api.routesForLocationHandler))
	mux.Handle("GET /api/where/schedule-for-stop/{id}", validateAPIKey(api, api.scheduleForStopHandler))
}
