package main

import (
	"encoding/json"
	"maglev.onebusaway.org/internal/models"
	"net/http"
)

func (app *application) sendResponse(w http.ResponseWriter, r *http.Request, response models.ResponseModel) {
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) sendNull(w http.ResponseWriter, r *http.Request) { // nolint:unused
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte("null"))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) sendNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)

	response := models.ResponseModel{
		Code:        http.StatusNotFound,
		CurrentTime: models.ResponseCurrentTime(),
		Text:        "resource not found",
		Version:     2,
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) sendUnauthorized(w http.ResponseWriter, r *http.Request) { // nolint:unused
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	response := models.ResponseModel{
		Code:        http.StatusUnauthorized,
		CurrentTime: models.ResponseCurrentTime(),
		Text:        "permission denied",
		Version:     1,
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
