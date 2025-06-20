package restapi

import (
	"context"
	"net/http"
	"sort"
	"time"

	"maglev.onebusaway.org/gtfsdb"
	"maglev.onebusaway.org/internal/models"
	"maglev.onebusaway.org/internal/utils"
)

func (api *RestAPI) tripDetailsHandler(w http.ResponseWriter, r *http.Request) {
	queryParamID := utils.ExtractIDFromParams(r)
	agencyID, tripID, err := utils.ExtractAgencyIDAndCodeID(queryParamID)
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}

	ctx := r.Context()

	trip, err := api.GtfsManager.GtfsDB.Queries.GetTrip(ctx, tripID)
	if err != nil {
		api.sendNotFound(w, r)
		return
	}

	route, err := api.GtfsManager.GtfsDB.Queries.GetRoute(ctx, trip.RouteID)
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}

	loc, _ := time.LoadLocation("America/Los_Angeles") // TODO: Get dynamically from agency if needed
	now := time.Now().In(loc)
	serviceDate := now.Truncate(24 * time.Hour)
	serviceDateMillis := serviceDate.Unix() * 1000

	nextTripID, previousTripID, err := api.GetNextAndPreviousTripIDs(ctx, &trip, tripID, agencyID, serviceDate)
	if err != nil {
		api.serverErrorResponse(w, r, err)
		return
	}

	// TODO: after adding service alerts data, implement this properly
	situationIDs := []string{}

	status, _ := api.buildTripStatus(agencyID, trip.ID, serviceDate)

	tripsToInclude := []string{utils.FormCombinedID(agencyID, trip.ID)}

	if nextTripID != "" {
		tripsToInclude = append(tripsToInclude, nextTripID)
	}
	if previousTripID != "" {
		tripsToInclude = append(tripsToInclude, previousTripID)
	}

	referencedTrips := []*models.Trip{}
	referencedStopTimes := []*models.StopTime{}

	for _, tripID := range tripsToInclude {

		_, refTripID, err := utils.ExtractAgencyIDAndCodeID(tripID)
		if err != nil {
			continue
		}

		if refTripID == trip.ID && len(referencedTrips) > 0 {
			continue
		}

		refTrip, err := api.GtfsManager.GtfsDB.Queries.GetTrip(ctx, refTripID)
		if err != nil {
			continue
		}

		refRoute, err := api.GtfsManager.GtfsDB.Queries.GetRoute(ctx, refTrip.RouteID)
		if err != nil {
			continue
		}

		var blockID string
		if refTrip.BlockID.Valid && refTrip.BlockID.String != "" {
			blockID = utils.FormCombinedID(agencyID, refTrip.BlockID.String)
		} else {
			blockID = ""
		}

		refTripModel := &models.Trip{
			ID:             tripID,
			RouteID:        utils.FormCombinedID(agencyID, refTrip.RouteID),
			ServiceID:      utils.FormCombinedID(agencyID, refTrip.ServiceID),
			ShapeID:        utils.FormCombinedID(agencyID, refTrip.ShapeID.String),
			TripHeadsign:   refTrip.TripHeadsign.String,
			TripShortName:  refTrip.TripShortName.String,
			DirectionID:    refTrip.DirectionID.Int64,
			BlockID:        blockID,
			RouteShortName: refRoute.ShortName.String,
			TimeZone:       "",
			PeakOffPeak:    0,
		}

		referencedTrips = append(referencedTrips, refTripModel)

		refStopTimes, err := api.GtfsManager.GtfsDB.Queries.GetStopTimesForTrip(ctx, refTripID)
		if err != nil {
			continue
		}

		for _, st := range refStopTimes {
			stopTimeModel := models.StopTime{
				ArrivalTime:         int(st.ArrivalTime),
				DepartureTime:       int(st.DepartureTime),
				StopID:              utils.FormCombinedID(agencyID, st.StopID),
				StopHeadsign:        st.StopHeadsign.String,
				DistanceAlongTrip:   st.ShapeDistTraveled.Float64,
				HistoricalOccupancy: "",
			}
			referencedStopTimes = append(referencedStopTimes, &stopTimeModel)
		}
	}

	stopTimesVals := make([]models.StopTime, len(referencedStopTimes))
	for i, st := range referencedStopTimes {
		if st != nil {
			stopTimesVals[i] = *st
		}
	}

	stopIDs := make([]string, len(stopTimesVals))
	for i, st := range stopTimesVals {
		stopIDs[i] = st.StopID
	}

	schedule := &models.Schedule{
		StopTimes:      stopTimesVals,
		TimeZone:       loc.String(),
		Frequency:      0,
		NextTripID:     nextTripID,
		PreviousTripID: previousTripID,
	}

	tripDetails := &models.TripDetails{
		TripID:       utils.FormCombinedID(agencyID, trip.ID),
		ServiceDate:  serviceDateMillis,
		Schedule:     schedule,
		Status:       status,
		Frequency:    nil,
		SituationIDs: situationIDs,
	}

	references := models.NewEmptyReferences()

	referencedTripsIface := make([]interface{}, len(referencedTrips))
	for i, t := range referencedTrips {
		referencedTripsIface[i] = t
	}
	references.Trips = referencedTripsIface

	routeModel := models.NewRoute(
		utils.FormCombinedID(agencyID, route.ID),
		agencyID,
		route.ShortName.String,
		route.LongName.String,
		route.Desc.String,
		models.RouteType(route.Type),
		route.Url.String,
		route.Color.String,
		route.TextColor.String,
		route.ShortName.String,
	)
	references.Routes = append(references.Routes, routeModel)

	agency, err := api.GtfsManager.GtfsDB.Queries.GetAgency(ctx, route.AgencyID)
	if err == nil {
		agencyModel := models.NewAgencyReference(
			agency.ID,
			agency.Name,
			agency.Url,
			agency.Timezone,
			agency.Lang.String,
			agency.Phone.String,
			agency.Email.String,
			agency.FareUrl.String,
			"",
			false,
		)
		references.Agencies = append(references.Agencies, agencyModel)
	}

	for _, stopID := range stopIDs {

		_, originalStopID, err := utils.ExtractAgencyIDAndCodeID(stopID)
		if err != nil {
			api.serverErrorResponse(w, r, err)
			return
		}

		stop, err := api.GtfsManager.GtfsDB.Queries.GetStop(ctx, originalStopID)
		if err != nil {
			continue
		}

		routesForStop, err := api.GtfsManager.GtfsDB.Queries.GetRoutesForStop(ctx, originalStopID)
		if err != nil {
			api.serverErrorResponse(w, r, err)
			return
		}

		combinedRouteIDs := make([]string, len(routesForStop))
		for i, rt := range routesForStop {
			combinedRouteIDs[i] = utils.FormCombinedID(agencyID, rt.ID)
		}

		stopModel := &models.Stop{
			ID:                 utils.FormCombinedID(agencyID, stop.ID),
			Name:               stop.Name.String,
			Lat:                stop.Lat,
			Lon:                stop.Lon,
			Code:               stop.Code.String,
			Direction:          "NE",                         // TODO
			LocationType:       int(stop.LocationType.Int64), // Cast
			WheelchairBoarding: "UNKNOWN",
			RouteIDs:           combinedRouteIDs,
			StaticRouteIDs:     combinedRouteIDs,
		}
		references.Stops = append(references.Stops, stopModel)
	}

	response := models.NewEntryResponse(tripDetails, references)
	api.sendResponse(w, r, response)
}

func (api *RestAPI) buildTripStatus(
	agencyID, tripID string,
	serviceDate time.Time,
) (*models.TripStatus, error) {
	vehicle := api.GtfsManager.GetVehicleForTrip(tripID)

	if vehicle == nil {
		return nil, nil
	}

	status := &models.TripStatus{
		ServiceDate:                serviceDate.Unix() * 1000,
		ActiveTripID:               utils.FormCombinedID(agencyID, tripID),
		Phase:                      "IN_PROGRESS", // TODO
		Status:                     "IN_PROGRESS", // TODO
		Predicted:                  true,
		VehicleID:                  utils.FormCombinedID(agencyID, vehicle.ID.ID),
		Position:                   models.Location{Lat: *vehicle.Position.Latitude, Lon: *vehicle.Position.Longitude},
		LastKnownLocation:          models.Location{Lat: *vehicle.Position.Latitude, Lon: *vehicle.Position.Longitude},
		Orientation:                float64(*vehicle.Position.Bearing),
		ScheduleDeviation:          int(*vehicle.Position.Bearing),
		DistanceAlongTrip:          0, // TODO:
		ScheduledDistanceAlongTrip: 0, // TODO
		TotalDistanceAlongTrip:     0, // TODO
		LastKnownDistanceAlongTrip: 0, // TODO
		LastUpdateTime:             vehicle.Timestamp.Unix() * 1000,
		LastLocationUpdateTime:     vehicle.Timestamp.Unix() * 1000,
		BlockTripSequence:          0,
		ClosestStop:                "",
		ClosestStopTimeOffset:      0,
		NextStop:                   "",
		NextStopTimeOffset:         0,
		OccupancyStatus:            vehicle.OccupancyStatus.String(),
		OccupancyCount:             0, // TODO
		OccupancyCapacity:          int(*vehicle.OccupancyPercentage),
		SituationIDs:               []string{}, // TODO
	}

	return status, nil
}

// retrieves the next and previous trip IDs for a given trip in the same block
func (api *RestAPI) GetNextAndPreviousTripIDs(ctx context.Context, trip *gtfsdb.Trip, tripID string, agencyID string, serviceDate time.Time) (nextTripID string, previousTripID string, err error) {
	if !trip.BlockID.Valid {
		return "", "", nil
	}

	blockTrips, err := api.GtfsManager.GtfsDB.Queries.GetTripsByBlockID(ctx, trip.BlockID)
	if err != nil {
		return "", "", err
	}

	if len(blockTrips) == 0 {
		return "", "", nil
	}

	type TripWithDetails struct {
		TripID    string
		StartTime int
		EndTime   int
		IsActive  bool
	}

	tripsWithDetails := []TripWithDetails{}

	for _, blockTrip := range blockTrips {
		isActive, err := api.GtfsManager.IsServiceActiveOnDate(ctx, blockTrip.ServiceID, serviceDate)
		if err != nil || isActive == 0 {
			continue
		}

		stopTimes, err := api.GtfsManager.GtfsDB.Queries.GetStopTimesForTrip(ctx, blockTrip.ID)
		if err != nil || len(stopTimes) == 0 {
			continue
		}

		startTime := int(^uint(0) >> 1) // max int value
		endTime := 0

		for _, st := range stopTimes {
			if st.DepartureTime > 0 && int(st.DepartureTime) < startTime {
				startTime = int(st.DepartureTime)
			}

			if st.ArrivalTime > 0 && int(st.ArrivalTime) > endTime {
				endTime = int(st.ArrivalTime)
			}
		}

		if startTime != int(^uint(0)>>1) && endTime > 0 {
			tripsWithDetails = append(tripsWithDetails, TripWithDetails{
				TripID:    blockTrip.ID,
				StartTime: startTime,
				EndTime:   endTime,
				IsActive:  isActive > 0,
			})
		}
	}

	sort.Slice(tripsWithDetails, func(i, j int) bool {
		if tripsWithDetails[i].IsActive && !tripsWithDetails[j].IsActive {
			return true
		}
		if !tripsWithDetails[i].IsActive && tripsWithDetails[j].IsActive {
			return false
		}
		return tripsWithDetails[i].StartTime < tripsWithDetails[j].StartTime
	})

	currentIndex := -1
	for i, t := range tripsWithDetails {
		if t.TripID == tripID {
			currentIndex = i
			break
		}
	}

	if currentIndex != -1 {
		if currentIndex > 0 {
			previousTripID = utils.FormCombinedID(agencyID, tripsWithDetails[currentIndex-1].TripID)
		}

		if currentIndex < len(tripsWithDetails)-1 {
			nextTripID = utils.FormCombinedID(agencyID, tripsWithDetails[currentIndex+1].TripID)
		}
	}

	return nextTripID, previousTripID, nil
}
