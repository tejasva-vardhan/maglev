package restapi

import (
	"net/http"
	"time"

	"maglev.onebusaway.org/internal/app"
)

type RestAPI struct {
	*app.Application
	rateLimiter func(http.Handler) http.Handler
}

// NewRestAPI creates a new RestAPI instance with initialized rate limiter
func NewRestAPI(app *app.Application) *RestAPI {
	return &RestAPI{
		Application: app,
		rateLimiter: NewRateLimitMiddleware(app.Config.RateLimit, time.Second),
	}
}
