package app

import "net/http"

func (app *Application) RequestHasInvalidAPIKey(r *http.Request) bool {
	key := r.URL.Query().Get("key")
	return app.IsInvalidAPIKey(key)
}

func (app *Application) IsInvalidAPIKey(key string) bool {
	if key == "" {
		return true
	}

	// For example, checking against keys stored in your app Config:
	validKeys := app.Config.ApiKeys
	for _, validKey := range validKeys {
		if key == validKey {
			return false
		}
	}

	return true
}
