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

func (app *application) sendNull(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte("null"))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
