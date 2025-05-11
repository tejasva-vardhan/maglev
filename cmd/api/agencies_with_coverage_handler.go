package main

import (
	"maglev.onebusaway.org/internal/models"
	"net/http"
)

func (api *restAPI) agenciesWithCoverageHandler(w http.ResponseWriter, r *http.Request) {
	agencies := api.app.GtfsManager.GetAgencies()
	lat, lon, latSpan, lonSpan := api.app.GtfsManager.GetRegionBounds()
	agenciesWithCoverage := make([]models.AgencyCoverage, 0, len(agencies))
	agencyReferences := make([]models.AgencyReference, 0, len(agencies))

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
