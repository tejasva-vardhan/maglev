package main

import (
	"encoding/json"
	"maglev.onebusaway.org/internal/models"
	"net/http"
)

// invalidAPIKeyResponse sends a 401 Unauthorized response with the required format
// for invalid API key errors
func (api *restAPI) invalidAPIKeyResponse(w http.ResponseWriter, r *http.Request) {
	// Create response with the specific format required
	response := struct {
		Code        int    `json:"code"`
		CurrentTime int64  `json:"currentTime"`
		Text        string `json:"text"`
		Version     int    `json:"version"`
	}{
		Code:        http.StatusUnauthorized,
		CurrentTime: models.ResponseCurrentTime(),
		Text:        "permission denied",
		Version:     1, // Note: This is version 1, not 2 as in a successful response. Probably a mistake, but back-compat.
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		api.app.Logger.Error("failed to encode invalid API key response", "error", err)
	}
}

func (api *restAPI) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	// Send a 500 Internal Server Error response
	response := struct {
		Code        int    `json:"code"`
		CurrentTime int64  `json:"currentTime"`
		Text        string `json:"text"`
		Version     int    `json:"version"`
	}{
		Code:        http.StatusInternalServerError,
		CurrentTime: models.ResponseCurrentTime(),
		Text:        "internal server error",
		Version:     1,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	encoderErr := json.NewEncoder(w).Encode(response)
	if encoderErr != nil {
		api.app.Logger.Error("failed to encode server error response", "error", encoderErr)
	}
}
