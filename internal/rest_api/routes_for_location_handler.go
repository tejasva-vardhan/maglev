package restapi

import (
	"context"
	"maglev.onebusaway.org/internal/models"
	"maglev.onebusaway.org/internal/utils"
	"net/http"
	"strings"
)

func (api *RestAPI) routesForLocationHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	lat, fieldErrors := utils.ParseFloatParam(queryParams, "lat", nil)
	lon, _ := utils.ParseFloatParam(queryParams, "lon", fieldErrors)
	radius, _ := utils.ParseFloatParam(queryParams, "radius", fieldErrors)
	latSpan, _ := utils.ParseFloatParam(queryParams, "latSpan", fieldErrors)
	lonSpan, _ := utils.ParseFloatParam(queryParams, "lonSpan", fieldErrors)
	query := queryParams.Get("query")
	query = strings.ToLower(query)
	if len(fieldErrors) > 0 {
		api.validationErrorResponse(w, r, fieldErrors)
		return
	}
	if radius == 0 {
		// Default radius to 600 meters if not specified
		radius = 600
		if query != "" {
			radius = 10000
		}
	}

	stops := api.GtfsManager.GetStopsForLocation(lat, lon, radius, latSpan, lonSpan, query, 50, true)

	ctx := context.Background()
	var results = []models.Route{}
	routeIDs := map[string]bool{}
	agencyIDs := map[string]bool{}

	for _, stop := range stops {
		routes, err := api.GtfsManager.GtfsDB.Queries.GetRoutesForStop(ctx, stop.Id)
		if err != nil || len(routes) == 0 {
			continue
		}

		for _, route := range routes {
			if query != "" && strings.ToLower(route.ShortName.String) != query {
				continue
			}
			agencyIDs[route.AgencyID] = true
			if !routeIDs[route.ID] {
				results = append(results, models.NewRoute(
					utils.FormCombinedID(route.AgencyID, route.ID),
					route.AgencyID,
					route.ShortName.String,
					route.LongName.String,
					route.Desc.String,
					models.RouteType(route.Type),
					route.Url.String,
					route.Color.String,
					route.TextColor.String,
					route.ShortName.String,
				))
			}
			routeIDs[route.ID] = true
		}
	}

	agencies := utils.FilterAgencies(api.GtfsManager.GetAgencies(), agencyIDs)

	references := models.ReferencesModel{
		Agencies:   agencies,
		Routes:     []interface{}{},
		Situations: []interface{}{},
		StopTimes:  []interface{}{},
		Stops:      []models.Stop{},
		Trips:      []interface{}{},
	}

	response := models.NewListResponseWithRange(results, references, len(results) == 0)
	api.sendResponse(w, r, response)
}
