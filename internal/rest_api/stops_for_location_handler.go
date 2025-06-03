package restapi

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jamespfennell/gtfs"
	"maglev.onebusaway.org/gtfsdb"
	"maglev.onebusaway.org/internal/models"
	"maglev.onebusaway.org/internal/utils"
)

func (api *RestAPI) stopsForLocationHandler(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	lat, fieldErrors := parseFloatParam(queryParams, "lat", nil)
	lon, _ := parseFloatParam(queryParams, "lon", fieldErrors)
	radius, _ := parseFloatParam(queryParams, "radius", fieldErrors)
	latSpan, _ := parseFloatParam(queryParams, "latSpan", fieldErrors)
	lonSpan, _ := parseFloatParam(queryParams, "lonSpan", fieldErrors)
	query := queryParams.Get("query")

	if len(fieldErrors) > 0 {
		api.validationErrorResponse(w, r, fieldErrors)
		return
	}

	stops := api.GtfsManager.GetStopsForLocation(lat, lon, radius, latSpan, lonSpan, query, 100)

	ctx := context.Background()
	var results []models.Stop
	routeIDs := map[string]bool{}
	agencyIDs := map[string]bool{}

	for _, stop := range stops {
		ridsIface, err := api.GtfsManager.GtfsDB.Queries.GetRouteIDsForStop(ctx, stop.Id)
		if err != nil || len(ridsIface) == 0 {
			continue
		}

		var rids []string
		for _, rid := range ridsIface {
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

	agencies := filterAgencies(api.GtfsManager.GetAgencies(), agencyIDs)
	routes := filterRoutes(api.GtfsManager.GtfsDB.Queries, ctx, routeIDs)

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
func parseFloatParam(params url.Values, key string, fieldErrors map[string][]string) (float64, map[string][]string) {
	if fieldErrors == nil {
		fieldErrors = make(map[string][]string)
	}

	val := params.Get(key)
	if val == "" {
		return 0, fieldErrors
	}

	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		fieldErrors[key] = append(fieldErrors[key], fmt.Sprintf("Invalid field value for field %q.", key))
	}
	return f, fieldErrors
}

func filterAgencies(all []gtfs.Agency, present map[string]bool) []models.AgencyReference {
	var refs []models.AgencyReference
	for _, a := range all {
		if present[a.Id] {
			refs = append(refs, models.NewAgencyReference(
				a.Id, a.Name, a.Url, a.Timezone, a.Language, a.Phone, a.Email, a.FareUrl, "", false,
			))
		}
	}
	return refs
}

func filterRoutes(q *gtfsdb.Queries, ctx context.Context, present map[string]bool) []interface{} {
	routes, err := q.ListRoutes(ctx)
	if err != nil {
		return nil
	}
	var refs []interface{}
	for _, r := range routes {
		if present[r.ID] {
			refs = append(refs, models.NewRoute(
				r.ID, r.AgencyID, r.ShortName.String, r.LongName.String,
				r.Desc.String, models.RouteType(r.Type), r.Url.String,
				r.Color.String, r.TextColor.String, r.ShortName.String,
			))
		}
	}
	return refs
}
