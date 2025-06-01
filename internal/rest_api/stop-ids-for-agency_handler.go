package restapi

import (
	"context"
	"net/http"

	"maglev.onebusaway.org/internal/models"
	"maglev.onebusaway.org/internal/utils"
)

func (api *RestAPI) stopIDsForAgencyHandler(w http.ResponseWriter, r *http.Request) {

	id := utils.ExtractIDFromParams(r)

	agency := api.GtfsManager.FindAgency(id)

	if agency == nil {
		api.sendNull(w, r)
		return
	}

	ctx := context.Background()

	stopIDs, err := api.GtfsManager.GtfsDB.Queries.GetStopIDsForAgency(ctx, id)

	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}

	response := make([]string, 0, len(stopIDs))
	for _, stopID := range stopIDs {
		response = append(response, utils.FormCombinedID(id, stopID))
	}

	api.sendResponse(w, r, models.NewListResponse(response, models.NewEmptyReferences()))

}
