package restapi

import (
	"net/http"

	"github.com/klauspost/compress/gzhttp"
)

// applyGzipMiddleware wraps a handler with gzip compression
func applyGzipMiddleware(next http.Handler) http.Handler {
	// Use klauspost/compress for better performance
	return gzhttp.GzipHandler(next)
}

// createCompressedAPIHandler creates an API handler with compression enabled
func createCompressedAPIHandler(api *RestAPI) http.Handler {
	// Create the base handler (this would normally be done in routes.go)
	mux := http.NewServeMux()
	
	// Add a simple test route for the integration test
	mux.HandleFunc("/api/where/agencies-with-coverage.json", func(w http.ResponseWriter, r *http.Request) {
		api.agenciesWithCoverageHandler(w, r)
	})
	
	// Apply compression middleware
	return applyGzipMiddleware(mux)
}