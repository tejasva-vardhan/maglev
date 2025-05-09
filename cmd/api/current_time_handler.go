package main

import (
	"maglev.onebusaway.org/internal/models"
	"net/http"
	"time"
)

// Declare a handler which writes a JSON response with information about the
// current time.
func (app *application) currentTimeHandler(w http.ResponseWriter, r *http.Request) {
	timeData := models.NewCurrentTimeData(time.Now())
	response := models.NewOKResponse(timeData)

	app.sendResponse(w, r, response)
}
