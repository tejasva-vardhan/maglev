package main

import (
	"github.com/julienschmidt/httprouter"
	"maglev.onebusaway.org/internal/models"
	"net/http"
	"strings"
)

func (app *application) agencyHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id := params.ByName("id.json")
	id = strings.Split(id, ".json")[0]
	agency := app.gtfsManager.FindAgency(id)
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
