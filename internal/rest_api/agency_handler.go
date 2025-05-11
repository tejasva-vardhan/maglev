package restapi

import (
	"maglev.onebusaway.org/internal/models"
	"maglev.onebusaway.org/internal/utils"
	"net/http"
)

func (api *RestAPI) agencyHandler(w http.ResponseWriter, r *http.Request) {
	id := utils.ExtractIDFromParams(r)
	agency := api.App.GtfsManager.FindAgency(id)

	if agency == nil {
		api.sendNotFound(w, r)
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
	api.sendResponse(w, r, response)
}
