package main

import (
	"maglev.onebusaway.org/internal/models"
	"maglev.onebusaway.org/internal/utils"
	"net/http"
)

func (app *application) agencyHandler(w http.ResponseWriter, r *http.Request) {
	if app.requestHasInvalidAPIKey(r) {
		app.invalidAPIKeyResponse(w, r)
		return
	}

	id := utils.ExtractIDFromParams(r)
	agency := app.gtfsManager.FindAgency(id)

	if agency == nil {
		app.sendNotFound(w, r)
		return
	}

	agencyData := models.NewAgencyReference(
		agency.Id,
		agency.Name,
		agency.Url,
		agency.Timezone,
		agency.Language,
		agency.Phone,
		agency.Email,
		agency.FareUrl,
		"",
		false,
	)

	response := models.NewEntryResponse(agencyData, models.NewEmptyReferences())
	app.sendResponse(w, r, response)
}
