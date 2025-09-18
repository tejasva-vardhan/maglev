package restapi

import (
	"context"
	"database/sql"
	"github.com/OneBusAway/go-gtfs"
	"maglev.onebusaway.org/gtfsdb"
	"maglev.onebusaway.org/internal/models"
	"maglev.onebusaway.org/internal/utils"
	"net/http"
	"time"
)

func (api *RestAPI) tripsForRouteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	agencyID, routeID, err := utils.ExtractAgencyIDAndCodeID(utils.ExtractIDFromParams(r))
	if err != nil {
		http.Error(w, "null", http.StatusBadRequest)
		return
	}

	includeSchedule := r.URL.Query().Get("includeSchedule") != "false"
	includeStatus := r.URL.Query().Get("includeStatus") != "false"

	currentAgency, err := api.GtfsManager.GtfsDB.Queries.GetAgency(ctx, agencyID)
	if err != nil {
		http.Error(w, "null", http.StatusNotFound)
		return
	}

	currentLocation, err := time.LoadLocation(currentAgency.Timezone)
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}

	timeParam := r.URL.Query().Get("time")
	formattedDate, currentTime, fieldErrors, success := utils.ParseTimeParameter(timeParam, currentLocation)
	if !success {
		api.validationErrorResponse(w, r, fieldErrors)
		return
	}

	serviceIDs, err := api.GtfsManager.GtfsDB.Queries.GetActiveServiceIDsForDate(ctx, formattedDate)
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}

	routeTrips, err := api.GtfsManager.GtfsDB.Queries.GetTripsForRouteInActiveServiceIDs(ctx, gtfsdb.GetTripsForRouteInActiveServiceIDsParams{
		RouteID:    routeID,
		ServiceIds: serviceIDs,
	})
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}

	blockIDs := make(map[string]struct{})
	for _, trip := range routeTrips {
		if trip.BlockID.String != "" {
			blockIDs[trip.BlockID.String] = struct{}{}
		}
	}

	var allRelatedTrips []gtfsdb.GetTripsByBlockIDRow
	for blockID := range blockIDs {
		relatedTrips, err := api.GtfsManager.GtfsDB.Queries.GetTripsByBlockID(ctx, sql.NullString{String: blockID, Valid: true})
		if err != nil {
			api.serverErrorResponse(w, r, err)
			return
		}
		for _, trip := range relatedTrips {
			if trip.RouteID != routeID {
				allRelatedTrips = append(allRelatedTrips, trip)
			}
		}
	}
	activeTrips := make(map[string]gtfs.Vehicle)
	realTimeVehicles := api.GtfsManager.GetRealTimeVehicles()

	for _, vehicle := range realTimeVehicles {
		if vehicle.Position == nil || vehicle.Trip == nil {
			continue
		}

		isOnRequestedRoute := vehicle.Trip.ID.RouteID == routeID
		isLinkedTrip := false
		for _, trip := range allRelatedTrips {
			if trip.ID == vehicle.Trip.ID.ID {
				isLinkedTrip = true
				break
			}
		}

		if isOnRequestedRoute || isLinkedTrip {
			activeTrips[vehicle.Trip.ID.ID] = vehicle
		}
	}

	allRoutes, allTrips, err := api.getAllRoutesAndTrips(ctx, w, r)
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}

	todayMidnight := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentLocation)
	tripAgencyResolver := NewTripAgencyResolver(allRoutes, allTrips)
	result := api.buildTripsForRouteEntries(ctx, activeTrips, tripAgencyResolver, includeSchedule, includeStatus, currentLocation, currentTime, todayMidnight, w, r)

	if result == nil {
		result = []models.TripsForRouteListEntry{}
	}

	references := buildTripReferences(api, w, r, ctx, includeSchedule, allRoutes, allTrips, nil, result)
	response := models.NewListResponseWithRange(result, references, false)
	api.sendResponse(w, r, response)
}

func (api *RestAPI) buildTripsForRouteEntries(
	ctx context.Context,
	activeTrips map[string]gtfs.Vehicle,
	tripAgencyResolver *TripAgencyResolver,
	includeSchedule bool,
	includeStatus bool,
	currentLocation *time.Location,
	currentTime time.Time,
	todayMidnight time.Time,
	w http.ResponseWriter,
	r *http.Request,
) []models.TripsForRouteListEntry {
	var result []models.TripsForRouteListEntry
	for _, vehicle := range activeTrips {
		pos := vehicle.Position
		if pos == nil {
			continue
		}

		tripID := vehicle.Trip.ID.ID
		agencyID := tripAgencyResolver.GetAgencyNameByTripID(tripID)
		var schedule *models.TripsSchedule
		var status *models.TripStatusForTripDetails
		if includeSchedule {
			schedule = api.buildScheduleForTrip(ctx, tripID, agencyID, currentTime, currentLocation, w, r)
			if schedule == nil {
				continue
			}
		}

		if includeStatus {
			status, _ = api.BuildTripStatus(ctx, agencyID, tripID, currentTime, currentTime)

			if status == nil {
				continue
			}
		}
		entry := models.TripsForRouteListEntry{
			Frequency:    nil,
			Schedule:     schedule,
			Status:       status,
			ServiceDate:  todayMidnight.UnixMilli(),
			SituationIds: api.GetSituationIDsForTrip(tripID),
			TripId:       utils.FormCombinedID(agencyID, tripID),
		}
		result = append(result, entry)
	}
	return result
}
func buildTripReferences[T interface{ GetTripId() string }](api *RestAPI, w http.ResponseWriter, r *http.Request, ctx context.Context, includeTrip bool, allRoutes []gtfsdb.Route, allTrips []gtfsdb.Trip, stops []*gtfs.Stop, trips []T) models.ReferencesModel {
	// Collect present trip IDs
	presentTrips := make(map[string]models.Trip, len(trips))
	presentRoutes := make(map[string]models.Route)
	for _, trip := range trips {
		_, tripID, _ := utils.ExtractAgencyIDAndCodeID(trip.GetTripId())
		presentTrips[tripID] = models.Trip{}
	}
	stopList := make([]models.Stop, 0, len(stops))
	for _, stop := range stops {
		if stop.Latitude == nil || stop.Longitude == nil {
			continue
		}
		routeIds, err := api.GtfsManager.GtfsDB.Queries.GetRouteIDsForStop(ctx, stop.Id)
		if err != nil {
			continue
		}

		routeIdsString := make([]string, len(routeIds))
		for i, id := range routeIds {
			presentRoutes[id.(string)] = models.Route{}
			routeIdsString[i] = id.(string)
		}

		stopList = append(stopList, models.Stop{
			Code:               stop.Code,
			Direction:          "NA", // TODO add direction
			ID:                 stop.Id,
			Lat:                *stop.Latitude,
			Lon:                *stop.Longitude,
			LocationType:       0,
			Name:               stop.Name,
			Parent:             "",
			RouteIDs:           routeIdsString,
			StaticRouteIDs:     routeIdsString,
			WheelchairBoarding: utils.MapWheelchairBoarding(stop.WheelchairBoarding),
		})
	}

	// Collect present routes and fill presentTrips with details
	for _, trip := range allTrips {
		if _, exists := presentTrips[trip.ID]; exists {
			presentTrips[trip.ID] = models.Trip{
				ID:            trip.ID,
				RouteID:       trip.RouteID,
				ServiceID:     trip.ServiceID,
				TripHeadsign:  trip.TripHeadsign.String,
				TripShortName: trip.TripShortName.String,
				DirectionID:   trip.DirectionID.Int64,
				BlockID:       trip.BlockID.String,
				ShapeID:       trip.ShapeID.String,
				PeakOffPeak:   0,
				TimeZone:      "",
			}
			presentRoutes[trip.RouteID] = models.Route{}
		}
	}

	// Collect agencies for present routes
	presentAgencies := make(map[string]models.AgencyReference)
	for _, route := range allRoutes {
		if _, exists := presentRoutes[route.ID]; exists {
			presentRoutes[route.ID] = models.NewRoute(
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
			)
			currentAgency, err := api.GtfsManager.GtfsDB.Queries.GetAgency(ctx, route.AgencyID)
			if err != nil {
				api.serverErrorResponse(w, r, err)
				return models.ReferencesModel{}
			}
			presentAgencies[currentAgency.ID] = models.NewAgencyReference(
				currentAgency.ID,
				currentAgency.Name,
				currentAgency.Url,
				currentAgency.Timezone,
				currentAgency.Lang.String,
				currentAgency.Phone.String,
				currentAgency.Email.String,
				currentAgency.FareUrl.String,
				"",
				false,
			)
		}
	}

	// Optionally include trip details
	tripsRefList := make([]interface{}, 0, len(presentTrips))
	if includeTrip {
		for _, trip := range presentTrips {
			tripDetails, err := api.GtfsManager.GtfsDB.Queries.GetTrip(ctx, trip.ID)
			if err == nil {
				var currentAgency = presentRoutes[tripDetails.RouteID].AgencyID
				tripsRefList = append(tripsRefList, models.Trip{
					ID:            utils.FormCombinedID(currentAgency, trip.ID),
					RouteID:       utils.FormCombinedID(currentAgency, tripDetails.RouteID),
					ServiceID:     utils.FormCombinedID(currentAgency, trip.ServiceID),
					TripHeadsign:  tripDetails.TripHeadsign.String,
					TripShortName: tripDetails.TripShortName.String,
					DirectionID:   tripDetails.DirectionID.Int64,
					BlockID:       tripDetails.BlockID.String,
					ShapeID:       utils.FormCombinedID(currentAgency, tripDetails.ShapeID.String),
					PeakOffPeak:   0,
					TimeZone:      "",
				})
			}
		}
	}

	// Convert presentRoutes and presentTrips maps to slices
	routes := make([]interface{}, 0, len(presentRoutes))
	for _, route := range presentRoutes {
		routes = append(routes, route)
	}

	agencyList := make([]models.AgencyReference, 0, len(presentAgencies))
	for _, agency := range presentAgencies {
		agencyList = append(agencyList, agency)
	}

	return models.ReferencesModel{
		Agencies:   agencyList,
		Routes:     routes,
		Situations: []interface{}{},
		StopTimes:  []interface{}{},
		Stops:      stopList,
		Trips:      tripsRefList,
	}
}
