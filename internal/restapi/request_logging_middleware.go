package restapi

import (
	"log/slog"
	"net/http"
	"time"

	"maglev.onebusaway.org/internal/logging"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// NewRequestLoggingMiddleware creates middleware that logs HTTP requests
func NewRequestLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Add logger to context for downstream handlers
			ctx := logging.WithLogger(r.Context(), logger)
			r = r.WithContext(ctx)

			// Wrap response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default status
			}

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Log the request
			duration := time.Since(start)

			logging.LogHTTPRequest(logger,
				r.Method,
				r.URL.Path, // Path without query parameters
				wrapped.statusCode,
				float64(duration.Nanoseconds())/1e6, // Convert to milliseconds
				slog.String("user_agent", r.Header.Get("User-Agent")),
				slog.String("component", "http_server"))
		})
	}
}

