package restapi

import (
	"context"
	"github.com/jamespfennell/gtfs"
	"github.com/twpayne/go-polyline"
	"maglev.onebusaway.org/gtfsdb"
	"maglev.onebusaway.org/internal/models"
	"maglev.onebusaway.org/internal/utils"
	"net/http"
	"time"
)

func (api *RestAPI) stopsForRouteHandler(w http.ResponseWriter, r *http.Request) {
	agencyID, routeID, _ := utils.ExtractAgencyIDAndCodeID(utils.ExtractIDFromParams(r))

	if routeID == "" || agencyID == "" {
		http.Error(w, "null", http.StatusInternalServerError)
		return
	}

	currentAgency := api.GtfsManager.FindAgency(agencyID)
	if currentAgency == nil {
		http.Error(w, "null", http.StatusInternalServerError)
		return
	}

	currentLocation, _ := time.LoadLocation(currentAgency.Timezone)
	currentTime := time.Now().In(currentLocation)
	formattedDate := currentTime.Format("20060102")

	ctx := context.Background()

	serviceIDs, _ := api.GtfsManager.GtfsDB.Queries.GetActiveServiceIDsForDate(ctx, formattedDate)

	allStops := make(map[string]bool)
	allPolylines := make([]models.Polyline, 0, 100)
	// Get trips for route that are active on the service date
	trips, err := api.GtfsManager.GtfsDB.Queries.GetTripsForRouteInActiveServiceIDs(ctx, gtfsdb.GetTripsForRouteInActiveServiceIDsParams{
		RouteID:    routeID,
		ServiceIds: serviceIDs,
	})
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}

	var stopGroupings []models.StopGrouping

	if len(trips) == 0 {
		// Fallback: get all trips for this route regardless of service date
		allTrips, err := api.GtfsManager.GtfsDB.Queries.GetAllTripsForRoute(ctx, routeID)
		if err != nil {
			api.serverErrorResponse(w, r, err)
			return
		}
		// Group trips by direction_id and trip_headsign
		processTripGroups(ctx, api, agencyID, routeID, allTrips, &stopGroupings, allStops, &allPolylines)
	} else {
		// Process trips for the current service date
		processTripGroups(ctx, api, agencyID, routeID, trips, &stopGroupings, allStops, &allPolylines)
	}

	allStopsIds := formatStopIDs(agencyID, allStops)
	stopsList := make([]models.Stop, 0, len(allStopsIds))
	for stopID, _ := range allStops {
		stop, _ := api.GtfsManager.GtfsDB.Queries.GetStop(ctx, stopID)
		routeIds, _ := api.GtfsManager.GtfsDB.Queries.GetRouteIDsForStop(ctx, stop.ID)
		// TODO calculate stop direction
		routeIdsString := make([]string, len(routeIds))
		for i, id := range routeIds {
			routeIdsString[i] = id.(string)
		}

		stopsList = append(stopsList, models.Stop{
			Code:               stop.Code.String,
			Direction:          "Direction",
			ID:                 utils.FormCombinedID(agencyID, stop.ID),
			Lat:                stop.Lat,
			LocationType:       int(stop.LocationType.Int64),
			Lon:                stop.Lon,
			Name:               stop.Name.String,
			RouteIDs:           routeIdsString,
			StaticRouteIDs:     routeIdsString,
			WheelchairBoarding: utils.MapWheelchairBoarding(gtfs.WheelchairBoarding(stop.WheelchairBoarding.Int64)),
		})
	}
	result := models.RouteEntry{
		Polylines:     allPolylines,
		RouteID:       utils.FormCombinedID(agencyID, routeID),
		StopGroupings: stopGroupings,
		StopIds:       allStopsIds,
	}

	agencyRef := models.NewAgencyReference(
		currentAgency.Id,
		currentAgency.Name,
		currentAgency.Url,
		currentAgency.Timezone,
		currentAgency.Language,
		currentAgency.Phone,
		currentAgency.Email,
		currentAgency.FareUrl,
		"",
		false,
	)

	references := models.ReferencesModel{
		Agencies:   []models.AgencyReference{agencyRef},
		Routes:     utils.GetAllRoutesRefs(api.GtfsManager.GtfsDB.Queries, ctx),
		Situations: []interface{}{},
		StopTimes:  []interface{}{},
		Stops:      stopsList,
		Trips:      []interface{}{},
	}

	response := models.NewEntryResponse(result, references)
	api.sendResponse(w, r, response)
}

func processTripGroups(
	ctx context.Context,
	api *RestAPI,
	agencyID string,
	routeID string,
	trips []gtfsdb.Trip,
	stopGroupings *[]models.StopGrouping,
	allStops map[string]bool,
	allPolylines *[]models.Polyline,
) {
	type directionHeadsignKey struct {
		DirectionID  int64
		TripHeadsign string
	}

	tripGroups := make(map[directionHeadsignKey][]gtfsdb.Trip)
	for _, trip := range trips {
		key := directionHeadsignKey{
			DirectionID:  trip.DirectionID.Int64,
			TripHeadsign: trip.TripHeadsign.String,
		}
		tripGroups[key] = append(tripGroups[key], trip)
	}

	for key, tripsInGroup := range tripGroups {
		repTrip := tripsInGroup[0]
		stopsList, err := api.GtfsManager.GtfsDB.Queries.GetOrderedStopIDsForTrip(ctx, repTrip.ID)
		stopIDs := make(map[string]bool)
		for _, stopID := range stopsList {
			stopIDs[stopID] = true
			allStops[stopID] = true
		}
		if err != nil {
			continue
		}
		shape, err := api.GtfsManager.GtfsDB.Queries.GetShapesGroupedByTripHeadSign(ctx,
			gtfsdb.GetShapesGroupedByTripHeadSignParams{
				RouteID:      routeID,
				TripHeadsign: repTrip.TripHeadsign,
			})

		polylines := generatePolylines(shape)
		*allPolylines = append(*allPolylines, polylines...)

		formattedStopIDs := formatStopIDs(agencyID, stopIDs)

		stopGroup := models.StopGroup{
			ID: utils.FormCombinedID(agencyID, routeID),
			Name: models.StopGroupName{
				Name:  key.TripHeadsign,
				Names: []string{key.TripHeadsign},
				Type:  "destination",
			},
			StopIds:   formattedStopIDs,
			Polylines: polylines,
		}

		*stopGroupings = append(*stopGroupings, models.StopGrouping{
			Ordered:    true,
			StopGroups: []models.StopGroup{stopGroup},
			Type:       "direction",
		})
	}
}

func generatePolylines(shapes []gtfsdb.GetShapesGroupedByTripHeadSignRow) []models.Polyline {
	var polylines []models.Polyline
	var coords [][]float64
	for _, shape := range shapes {
		coords = append(coords, []float64{shape.Lat, shape.Lon})
	}
	encodedPoints := polyline.EncodeCoords(coords)
	polylines = append(polylines, models.Polyline{
		Length: len(shapes),
		Levels: "",
		Points: string(encodedPoints),
	})
	return polylines
}

func formatStopIDs(agencyID string, stops map[string]bool) []string {
	var stopIDs []string
	for key, _ := range stops {
		stopID := utils.FormCombinedID(agencyID, key)
		stopIDs = append(stopIDs, stopID)
	}
	return stopIDs
}
