package main

import (
	"encoding/json"
	"maglev.onebusaway.org/internal/models"
	"net/http"
)

func (app *Application) sendResponse(w http.ResponseWriter, r *http.Request, response models.ResponseModel) {
	w.Header().Set("Content-Type", "Application/json")

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *Application) sendNull(w http.ResponseWriter, r *http.Request) { // nolint:unused
	w.Header().Set("Content-Type", "Application/json")
	_, err := w.Write([]byte("null"))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *Application) sendNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "Application/json")
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

func (app *Application) sendUnauthorized(w http.ResponseWriter, r *http.Request) { // nolint:unused
	w.Header().Set("Content-Type", "Application/json")
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
