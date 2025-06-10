package restapi

import (
	"context"
	"maglev.onebusaway.org/internal/models"
	"maglev.onebusaway.org/internal/utils"
	"net/http"
)

func (api *RestAPI) stopsForLocationHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	lat, fieldErrors := utils.ParseFloatParam(queryParams, "lat", nil)
	lon, _ := utils.ParseFloatParam(queryParams, "lon", fieldErrors)
	radius, _ := utils.ParseFloatParam(queryParams, "radius", fieldErrors)
	latSpan, _ := utils.ParseFloatParam(queryParams, "latSpan", fieldErrors)
	lonSpan, _ := utils.ParseFloatParam(queryParams, "lonSpan", fieldErrors)
	query := queryParams.Get("query")

	if len(fieldErrors) > 0 {
		api.validationErrorResponse(w, r, fieldErrors)
		return
	}

	stops := api.GtfsManager.GetStopsForLocation(lat, lon, radius, latSpan, lonSpan, query, 100, false)

	ctx := context.Background()
	var results []models.Stop
	routeIDs := map[string]bool{}
	agencyIDs := map[string]bool{}

	for _, stop := range stops {
		routeIds, err := api.GtfsManager.GtfsDB.Queries.GetRouteIDsForStop(ctx, stop.Id)
		if err != nil || len(routeIds) == 0 {
			continue
		}

		var rids []string
		for _, rid := range routeIds {
			ridStr, ok := rid.(string)
			if !ok {
				continue
			}
			agencyId, routeId, _ := utils.ExtractAgencyIDAndCodeID(ridStr)
			agencyIDs[agencyId] = true
			routeIDs[routeId] = true
			rids = append(rids, ridStr)
		}
		agency, err := api.GtfsManager.GtfsDB.Queries.GetAgencyForStop(ctx, stop.Id)

		if err != nil {
			continue
		}

		results = append(results, models.NewStop(
			stop.Id,
			"Direction",
			utils.FormCombinedID(agency.ID, stop.Id),
			stop.Name,
			"",
			utils.MapWheelchairBoarding(stop.WheelchairBoarding),
			*stop.Latitude,
			*stop.Longitude,
			0,
			rids,
			rids,
		))
	}

	agencies := utils.FilterAgencies(api.GtfsManager.GetAgencies(), agencyIDs)
	routes := utils.FilterRoutes(api.GtfsManager.GtfsDB.Queries, ctx, routeIDs)

	references := models.ReferencesModel{
		Agencies:   agencies,
		Routes:     routes,
		Situations: []interface{}{},
		StopTimes:  []interface{}{},
		Stops:      []interface{}{},
		Trips:      []interface{}{},
	}

	response := models.NewListResponseWithRange(results, references, len(results) == 0)
	api.sendResponse(w, r, response)
}
