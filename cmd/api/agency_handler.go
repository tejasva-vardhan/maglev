package main

import (
	"maglev.onebusaway.org/internal/models"
	"net/http"
	"strings"
)

func (app *application) agencyHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	id = strings.Split(id, ".json")[0]
	agency := app.gtfsManager.FindAgency(id)
	
	if agency == nil {
		app.sendNull(w, r)
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
