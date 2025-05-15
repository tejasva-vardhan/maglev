package restapi

import (
	"context"
	"maglev.onebusaway.org/internal/models"
	"net/http"
)

func (api *RestAPI) agenciesWithCoverageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	agencies, err := api.GtfsManager.GtfsDB.QueryAgencies(ctx)
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}
	
	lat, lon, latSpan, lonSpan := api.GtfsManager.GetRegionBounds()
	var agenciesWithCoverage []models.AgencyCoverage
	var agencyReferences []models.AgencyReference

	for _, a := range agencies {
		agenciesWithCoverage = append(
			agenciesWithCoverage,
			models.NewAgencyCoverage(a.Id, lat, latSpan, lon, lonSpan),
		)

		agencyReferences = append(
			agencyReferences,
			models.NewAgencyReference(
				a.Id,
				a.Name,
				a.Url,
				a.Timezone,
				a.Language,
				a.Phone,
				a.Email,
				a.FareUrl,
				"",
				false,
			),
		)
	}

	// Create references with the agency
	references := models.ReferencesModel{
		Agencies:   agencyReferences,
		Routes:     []interface{}{},
		Situations: []interface{}{},
		StopTimes:  []interface{}{},
		Stops:      []interface{}{},
		Trips:      []interface{}{},
	}

	response := models.NewListResponse(agenciesWithCoverage, references)
	api.sendResponse(w, r, response)
}
