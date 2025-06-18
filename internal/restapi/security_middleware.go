package restapi

import (
	"net/http"
)

// WithSecurityHeaders wraps the given handler with security headers middleware
func (api *RestAPI) WithSecurityHeaders(handler http.Handler) http.Handler {
	return securityHeaders(handler)
}

// securityHeaders adds essential security headers to all HTTP responses
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking attacks
		w.Header().Set("X-Frame-Options", "DENY")

		// Force HTTPS in production (browser will refuse HTTP connections)
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Prevent XSS attacks
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Control referrer information
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy - restrictive but allows API responses
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none';")

		// CORS headers for API access
		origin := r.Header.Get("Origin")
		if origin != "" {
			// Allow all origins for public transit API, but be explicit about it
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
		}

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
