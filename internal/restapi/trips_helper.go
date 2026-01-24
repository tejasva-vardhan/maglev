package restapi

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/OneBusAway/go-gtfs"
	"maglev.onebusaway.org/gtfsdb"
	"maglev.onebusaway.org/internal/models"
	"maglev.onebusaway.org/internal/utils"
)

// IMPORTANT: Caller must hold manager.RLock() before calling this method.
func (api *RestAPI) BuildTripStatus(
	ctx context.Context,
	agencyID, tripID string,
	serviceDate time.Time,
	currentTime time.Time,
) (*models.TripStatusForTripDetails, error) {
	status := &models.TripStatusForTripDetails{
		ActiveTripID:      utils.FormCombinedID(agencyID, tripID),
		ServiceDate:       serviceDate.Unix() * 1000,
		SituationIDs:      api.GetSituationIDsForTrip(ctx, tripID),
		OccupancyCapacity: -1,
		OccupancyCount:    -1,
	}

	vehicle := api.GtfsManager.GetVehicleForTrip(tripID)

	if vehicle != nil {
		if vehicle.ID != nil {
			status.VehicleID = vehicle.ID.ID
		}
		if vehicle.OccupancyStatus != nil {
			status.OccupancyStatus = vehicle.OccupancyStatus.String()
		}
		api.BuildVehicleStatus(ctx, vehicle, tripID, agencyID, status)
	}

	scheduleDeviation := api.GetScheduleDeviation(tripID)
	if scheduleDeviation != 0 {
		status.ScheduleDeviation = scheduleDeviation
		status.Predicted = true
	}

	_, activeTripRawID, err := utils.ExtractAgencyIDAndCodeID(status.ActiveTripID)
	if err != nil {
		return status, err
	}

	stopTimes, err := api.GtfsManager.GtfsDB.Queries.GetStopTimesForTrip(ctx, activeTripRawID)
	if err == nil && len(stopTimes) > 0 {
		stopTimesPtrs := make([]*gtfsdb.StopTime, len(stopTimes))
		for i := range stopTimes {
			stopTimesPtrs[i] = &stopTimes[i]
		}

		var closestStopID, nextStopID string
		var closestOffset, nextOffset int

		if vehicle != nil && vehicle.Position != nil {
			if vehicle.StopID != nil && *vehicle.StopID != "" {
				closestStopID = *vehicle.StopID
				closestOffset = api.calculateOffsetForStop(closestStopID, stopTimesPtrs, currentTime, serviceDate, scheduleDeviation)
				nextStopID, nextOffset = api.findNextStopAfter(closestStopID, stopTimesPtrs, currentTime, serviceDate, scheduleDeviation)
			} else if vehicle.CurrentStopSequence != nil {
				closestStopID, closestOffset = api.findClosestStopBySequence(
					stopTimesPtrs, *vehicle.CurrentStopSequence, currentTime, serviceDate, scheduleDeviation, vehicle,
				)
				nextStopID, nextOffset = api.findNextStopBySequence(
					ctx, stopTimesPtrs, *vehicle.CurrentStopSequence, currentTime, serviceDate, scheduleDeviation, vehicle, tripID, serviceDate,
				)
			} else {
				closestStopID, closestOffset, nextStopID, nextOffset = api.findStopsByScheduleDeviation(
					stopTimesPtrs, currentTime, serviceDate, scheduleDeviation,
				)
			}
		} else {
			stopDelays := api.GetStopDelaysFromTripUpdates(tripID)
			closestStopID, closestOffset = findClosestStopByTimeWithDelays(currentTime, serviceDate, stopTimesPtrs, stopDelays)
			nextStopID, nextOffset = findNextStopByTimeWithDelays(currentTime, serviceDate, stopTimesPtrs, stopDelays)
		}

		if closestStopID != "" {
			status.ClosestStop = utils.FormCombinedID(agencyID, closestStopID)
			status.ClosestStopTimeOffset = closestOffset
		}
		if nextStopID != "" {
			status.NextStop = utils.FormCombinedID(agencyID, nextStopID)
			status.NextStopTimeOffset = nextOffset
		}
	}

	if status.ClosestStop == "" || status.NextStop == "" {
		api.fillStopsFromSchedule(ctx, status, tripID, currentTime, serviceDate, agencyID)
	}

	shapeRows, shapeErr := api.GtfsManager.GtfsDB.Queries.GetShapePointsByTripID(ctx, tripID)
	if shapeErr == nil && len(shapeRows) > 1 {
		shapePoints := make([]gtfs.ShapePoint, len(shapeRows))
		for i, sp := range shapeRows {
			shapePoints[i] = gtfs.ShapePoint{
				Latitude:  sp.Lat,
				Longitude: sp.Lon,
			}
		}
		cumulativeDistances := preCalculateCumulativeDistances(shapePoints)
		status.TotalDistanceAlongTrip = cumulativeDistances[len(cumulativeDistances)-1]

		if vehicle != nil && vehicle.Position != nil && vehicle.Position.Latitude != nil && vehicle.Position.Longitude != nil {
			actualDistance := api.getVehicleDistanceAlongShapeContextual(ctx, tripID, vehicle)
			status.DistanceAlongTrip = actualDistance

			if scheduleDeviation != 0 && err == nil {
				scheduledDistance := api.calculateEffectiveDistanceAlongTrip(
					ctx, actualDistance, scheduleDeviation, currentTime, serviceDate,
					stopTimes, shapePoints, cumulativeDistances,
				)
				status.ScheduledDistanceAlongTrip = scheduledDistance
			}
		}
	}

	blockTripSequence := api.calculateBlockTripSequence(ctx, tripID, serviceDate)
	if blockTripSequence > 0 {
		status.BlockTripSequence = blockTripSequence
	}

	return status, nil
}

// IMPORTANT: Caller must hold manager.RLock() before calling this method.
func (api *RestAPI) BuildTripSchedule(ctx context.Context, agencyID string, serviceDate time.Time, trip *gtfsdb.Trip, loc *time.Location) (*models.Schedule, error) {
	stopTimes, err := api.GtfsManager.GtfsDB.Queries.GetStopTimesForTrip(ctx, trip.ID)
	if err != nil {
		return nil, err
	}

	shapeRows, err := api.GtfsManager.GtfsDB.Queries.GetShapePointsByTripID(ctx, trip.ID)
	var shapePoints []gtfs.ShapePoint
	if err == nil && len(shapeRows) > 0 {
		shapePoints = make([]gtfs.ShapePoint, len(shapeRows))
		for i, sp := range shapeRows {
			shapePoints[i] = gtfs.ShapePoint{
				Latitude:  sp.Lat,
				Longitude: sp.Lon,
			}
		}
	}

	var nextTripID, previousTripID string
	nextTripID, previousTripID, _, err = api.GetNextAndPreviousTripIDs(ctx, trip, agencyID, serviceDate)
	if err != nil {
		return nil, err
	}

	// Batch-fetch all stop coordinates at once
	stopIDs := make([]string, len(stopTimes))
	for i, st := range stopTimes {
		stopIDs[i] = st.StopID
	}

	stops, err := api.GtfsManager.GtfsDB.Queries.GetStopsByIDs(ctx, stopIDs)
	if err != nil {
		return nil, err
	}

	stopCoords := make(map[string]struct{ lat, lon float64 })
	for _, stop := range stops {
		stopCoords[stop.ID] = struct{ lat, lon float64 }{lat: stop.Lat, lon: stop.Lon}
	}

	stopTimesVals := api.calculateBatchStopDistances(stopTimes, shapePoints, stopCoords, agencyID)

	return &models.Schedule{
		StopTimes:      stopTimesVals,
		TimeZone:       loc.String(),
		Frequency:      0,
		NextTripID:     nextTripID,
		PreviousTripID: previousTripID,
	}, nil
}

// IMPORTANT: Caller must hold manager.RLock() before calling this method.
func (api *RestAPI) GetNextAndPreviousTripIDs(ctx context.Context, trip *gtfsdb.Trip, agencyID string, serviceDate time.Time) (nextTripID string, previousTripID string, stopTimes []gtfsdb.StopTime, err error) {
	if !trip.BlockID.Valid {
		return "", "", nil, nil
	}

	orderedTrips, err := api.GtfsManager.GtfsDB.Queries.GetTripsByBlockIDOrdered(ctx, gtfsdb.GetTripsByBlockIDOrderedParams{
		BlockID:    trip.BlockID,
		ServiceIds: []string{trip.ServiceID},
	})
	if err != nil {
		return "", "", nil, err
	}
	if len(orderedTrips) == 0 {
		return "", "", nil, nil
	}

	currentIndex := -1
	for i, t := range orderedTrips {
		if t.ID == trip.ID {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return "", "", nil, nil
	}

	if currentIndex > 0 {
		previousTripID = utils.FormCombinedID(agencyID, orderedTrips[currentIndex-1].ID)
	}

	if currentIndex < len(orderedTrips)-1 {
		nextTripID = utils.FormCombinedID(agencyID, orderedTrips[currentIndex+1].ID)
	}

	stopTimes, err = api.GtfsManager.GtfsDB.Queries.GetStopTimesForTrip(ctx, trip.ID)
	if err != nil {
		return nextTripID, previousTripID, nil, err
	}

	return nextTripID, previousTripID, stopTimes, nil
}

func (api *RestAPI) fillStopsFromSchedule(ctx context.Context, status *models.TripStatusForTripDetails, tripID string, currentTime time.Time, serviceDate time.Time, agencyID string) {
	stopTimes, err := api.GtfsManager.GtfsDB.Queries.GetStopTimesForTrip(ctx, tripID)
	if err != nil || len(stopTimes) == 0 {
		return
	}

	currentSeconds := utils.CalculateSecondsSinceServiceDate(currentTime, serviceDate)

	for i, st := range stopTimes {
		arrivalTime := st.ArrivalTime / 1e9
		if arrivalTime == 0 {
			arrivalTime = st.DepartureTime / 1e9
		}

		predictedArrival := arrivalTime + int64(status.ScheduleDeviation)

		if predictedArrival > currentSeconds {
			if i > 0 {
				status.ClosestStop = utils.FormCombinedID(agencyID, stopTimes[i-1].StopID)
				closestArrival := stopTimes[i-1].ArrivalTime / 1e9
				if closestArrival == 0 {
					closestArrival = stopTimes[i-1].DepartureTime / 1e9
				}
				status.ClosestStopTimeOffset = int(closestArrival + int64(status.ScheduleDeviation) - currentSeconds)
			}
			status.NextStop = utils.FormCombinedID(agencyID, st.StopID)
			status.NextStopTimeOffset = int(predictedArrival - currentSeconds)
			return
		}
	}

	if len(stopTimes) > 0 {
		lastStop := stopTimes[len(stopTimes)-1]
		status.ClosestStop = utils.FormCombinedID(agencyID, lastStop.StopID)
		arrivalTime := lastStop.ArrivalTime / 1e9
		if arrivalTime == 0 {
			arrivalTime = lastStop.DepartureTime / 1e9
		}
		status.ClosestStopTimeOffset = int(arrivalTime + int64(status.ScheduleDeviation) - currentSeconds)
	}
}

func findClosestStopByTimeWithDelays(currentTime time.Time, serviceDate time.Time, stopTimes []*gtfsdb.StopTime, stopDelays map[string]StopDelayInfo) (stopID string, offset int) {
	currentTimeSeconds := utils.CalculateSecondsSinceServiceDate(currentTime, serviceDate)
	var minTimeDiff int64 = math.MaxInt64
	var closestStopTimeSeconds int64

	for _, st := range stopTimes {
		var stopTimeSeconds int64
		if st.DepartureTime > 0 {
			stopTimeSeconds = st.DepartureTime / 1e9
		} else if st.ArrivalTime > 0 {
			stopTimeSeconds = st.ArrivalTime / 1e9
		} else {
			continue
		}

		if stopDelays != nil {
			if delayInfo, exists := stopDelays[st.StopID]; exists {
				if st.DepartureTime > 0 && delayInfo.DepartureDelay != 0 {
					stopTimeSeconds += delayInfo.DepartureDelay
				} else if delayInfo.ArrivalDelay != 0 {
					stopTimeSeconds += delayInfo.ArrivalDelay
				}
			}
		}

		timeDiff := int64(math.Abs(float64(currentTimeSeconds - stopTimeSeconds)))
		if timeDiff < minTimeDiff {
			minTimeDiff = timeDiff
			stopID = st.StopID
			closestStopTimeSeconds = stopTimeSeconds
		}
	}

	if stopID != "" {
		offset = int(closestStopTimeSeconds - currentTimeSeconds)
	}

	return
}

func findNextStopByTimeWithDelays(currentTime time.Time, serviceDate time.Time, stopTimes []*gtfsdb.StopTime, stopDelays map[string]StopDelayInfo) (stopID string, offset int) {
	currentTimeSeconds := utils.CalculateSecondsSinceServiceDate(currentTime, serviceDate)
	var minTimeDiff int64 = math.MaxInt64
	var nextStopTimeSeconds int64

	for _, st := range stopTimes {
		var stopTimeSeconds int64
		if st.DepartureTime > 0 {
			stopTimeSeconds = st.DepartureTime / 1e9
		} else if st.ArrivalTime > 0 {
			stopTimeSeconds = st.ArrivalTime / 1e9
		} else {
			continue
		}

		if stopDelays != nil {
			if delayInfo, exists := stopDelays[st.StopID]; exists {
				if st.DepartureTime > 0 && delayInfo.DepartureDelay != 0 {
					stopTimeSeconds += delayInfo.DepartureDelay
				} else if delayInfo.ArrivalDelay != 0 {
					stopTimeSeconds += delayInfo.ArrivalDelay
				}
			}
		}

		if stopTimeSeconds > currentTimeSeconds {
			timeDiff := stopTimeSeconds - currentTimeSeconds
			if timeDiff < minTimeDiff {
				minTimeDiff = timeDiff
				stopID = st.StopID
				nextStopTimeSeconds = stopTimeSeconds
			}
		}
	}

	if stopID != "" {
		offset = int(nextStopTimeSeconds - currentTimeSeconds)
	}

	return
}

func getDistanceAlongShape(lat, lon float64, shape []gtfs.ShapePoint) float64 {
	if len(shape) == 0 {
		return 0
	}
	var total float64
	var closestIndex int
	var minDist = math.MaxFloat64

	for i := range shape {
		dist := utils.Distance(lat, lon, shape[i].Latitude, shape[i].Longitude)
		if dist < minDist {
			minDist = dist
			closestIndex = i
		}
	}

	for i := 1; i <= closestIndex; i++ {
		total += utils.Distance(shape[i-1].Latitude, shape[i-1].Longitude, shape[i].Latitude, shape[i].Longitude)
	}

	return total
}

func getDistanceAlongShapeInRange(lat, lon float64, shape []gtfs.ShapePoint, minDistTraveled, maxDistTraveled float64) float64 {
	if len(shape) == 0 {
		return 0
	}

	cumulativeDistances := preCalculateCumulativeDistances(shape)
	var bestDist float64
	var minPointDist = math.MaxFloat64

	// If maxDistTraveled is 0, it might mean it wasn't provided or it's the start.
	// We should probably use the whole shape if the range is invalid.
	useRange := maxDistTraveled > minDistTraveled

	for i := 0; i < len(shape); i++ {
		if useRange && (cumulativeDistances[i] < minDistTraveled-10 || cumulativeDistances[i] > maxDistTraveled+10) {
			continue
		}

		dist := utils.Distance(lat, lon, shape[i].Latitude, shape[i].Longitude)
		if dist < minPointDist {
			minPointDist = dist
			bestDist = cumulativeDistances[i]
		}
	}

	return bestDist
}

func (api *RestAPI) setBlockTripSequence(ctx context.Context, tripID string, serviceDate time.Time, status *models.TripStatusForTripDetails) int {
	return api.calculateBlockTripSequence(ctx, tripID, serviceDate)
}

// calculateBlockTripSequence calculates the index of a trip within its block's ordered trip sequence
// for trips that are active on the given service date.
// Uses GetTripsByBlockIDOrdered to perform a single SQL JOIN instead of N+1 queries.
func (api *RestAPI) calculateBlockTripSequence(ctx context.Context, tripID string, serviceDate time.Time) int {
	blockID, err := api.GtfsManager.GtfsDB.Queries.GetBlockIDByTripID(ctx, tripID)

	if err != nil || !blockID.Valid || blockID.String == "" {
		return 0
	}

	blockTrips, err := api.GtfsManager.GtfsDB.Queries.GetTripsByBlockID(ctx, blockID)
	if err != nil || len(blockTrips) == 0 {
		return 0
	}

	tripIDs := make([]string, len(blockTrips))
	for i, bt := range blockTrips {
		tripIDs[i] = bt.ID
	}

	allStopTimes, err := api.GtfsManager.GtfsDB.Queries.GetStopTimesForTripIDs(ctx, tripIDs)
	if err != nil {
		return 0
	}

	stopTimesByTrip := make(map[string][]gtfsdb.StopTime)
	for _, st := range allStopTimes {
		stopTimesByTrip[st.TripID] = append(stopTimesByTrip[st.TripID], st)
	}

	type TripWithDetails struct {
		TripID    string
		StartTime int
	}

	activeTrips := []TripWithDetails{}

	for _, blockTrip := range blockTrips {
		isActive, err := api.GtfsManager.IsServiceActiveOnDate(ctx, blockTrip.ServiceID, serviceDate)
		if err != nil || isActive == 0 {
			continue
		}

		stopTimes, exists := stopTimesByTrip[blockTrip.ID]
		if !exists || len(stopTimes) == 0 {
			continue
		}

		startTime := math.MaxInt
		for _, st := range stopTimes {
			if st.DepartureTime > 0 && int(st.DepartureTime) < startTime {
				startTime = int(st.DepartureTime)
			}
		}

		if startTime != math.MaxInt {
			activeTrips = append(activeTrips, TripWithDetails{
				TripID:    blockTrip.ID,
				StartTime: startTime,
			})
		}
	}

	// Third, sort trips by start time, then by trip ID for deterministic ordering
	sort.Slice(activeTrips, func(i, j int) bool {
		if activeTrips[i].StartTime != activeTrips[j].StartTime {
			return activeTrips[i].StartTime < activeTrips[j].StartTime
		}
		return activeTrips[i].TripID < activeTrips[j].TripID
	})

	for i, trip := range activeTrips {
		if trip.TripID == tripID {
			return i
		}
	}
	return 0
}

func (api *RestAPI) calculateScheduleDeviationFromTripUpdates(
	tripID string,
) int {
	tripUpdates := api.GtfsManager.GetTripUpdatesForTrip(tripID)
	if len(tripUpdates) == 0 {
		return 0
	}

	tripUpdate := tripUpdates[0]

	if tripUpdate.Delay != nil {
		return int(tripUpdate.Delay.Seconds())
	}

	var bestDeviation int64 = 0
	var foundRelevantUpdate bool

	for _, stopTimeUpdate := range tripUpdate.StopTimeUpdates {
		if stopTimeUpdate.Arrival != nil && stopTimeUpdate.Arrival.Delay != nil {
			bestDeviation = int64(*stopTimeUpdate.Arrival.Delay / 1e9)
			foundRelevantUpdate = true
		} else if stopTimeUpdate.Departure != nil && stopTimeUpdate.Departure.Delay != nil {
			bestDeviation = int64(*stopTimeUpdate.Departure.Delay / 1e9)
			foundRelevantUpdate = true
		}

		if foundRelevantUpdate {
			break
		}
	}

	return int(bestDeviation)
}

func (api *RestAPI) getStopDelaysFromTripUpdates(tripID string) map[string]StopDelayInfo {
	delays := make(map[string]StopDelayInfo)

	tripUpdates := api.GtfsManager.GetTripUpdatesForTrip(tripID)
	if len(tripUpdates) == 0 {
		return delays
	}

	tripUpdate := tripUpdates[0]

	for _, stopTimeUpdate := range tripUpdate.StopTimeUpdates {
		if stopTimeUpdate.StopID == nil {
			continue
		}

		info := StopDelayInfo{}
		if stopTimeUpdate.Arrival != nil && stopTimeUpdate.Arrival.Delay != nil {
			info.ArrivalDelay = int64(stopTimeUpdate.Arrival.Delay.Seconds())
		}
		if stopTimeUpdate.Departure != nil && stopTimeUpdate.Departure.Delay != nil {
			info.DepartureDelay = int64(stopTimeUpdate.Departure.Delay.Seconds())
		}

		// Only add if we have at least one delay value
		if info.ArrivalDelay != 0 || info.DepartureDelay != 0 {
			delays[*stopTimeUpdate.StopID] = info
		}
	}

	return delays
}

// calculatePreciseDistanceAlongTripWithCoords calculates the distance along a trip's shape to a stop
// This optimized version accepts pre-calculated cumulative distances and stop coordinates
func (api *RestAPI) calculatePreciseDistanceAlongTripWithCoords(
	stopLat, stopLon float64,
	shapePoints []gtfs.ShapePoint,
	cumulativeDistances []float64,
) float64 {
	// Validate inputs
	if len(shapePoints) < 2 {
		return 0.0
	}

	// Validate that cumulative distances array matches shape points
	if len(cumulativeDistances) != len(shapePoints) {
		return 0.0
	}

	// Find the closest point on the shape to this stop
	var minDistance = math.Inf(1)
	var closestSegmentIndex int
	var projectionRatio float64

	for i := 0; i < len(shapePoints)-1; i++ {
		// Calculate distance from stop to this line segment
		distance, ratio := distanceToLineSegment(
			stopLat, stopLon,
			shapePoints[i].Latitude, shapePoints[i].Longitude,
			shapePoints[i+1].Latitude, shapePoints[i+1].Longitude,
		)

		if distance < minDistance {
			minDistance = distance
			closestSegmentIndex = i
			projectionRatio = ratio
		}
	}

	// Get cumulative distance to the start of the closest segment
	// cumulativeDistances[i] represents the total distance from start to point i
	cumulativeDistance := cumulativeDistances[closestSegmentIndex]

	// Add the projection distance within the closest segment
	if closestSegmentIndex < len(shapePoints)-1 {
		segmentDistance := utils.Distance(
			shapePoints[closestSegmentIndex].Latitude, shapePoints[closestSegmentIndex].Longitude,
			shapePoints[closestSegmentIndex+1].Latitude, shapePoints[closestSegmentIndex+1].Longitude,
		)
		cumulativeDistance += segmentDistance * projectionRatio
	}

	return cumulativeDistance
}

// calculatePreciseDistanceAlongTrip is the legacy version that fetches stop coordinates from the database
// Deprecated: Use calculatePreciseDistanceAlongTripWithCoords with batch-fetched coordinates instead
func (api *RestAPI) calculatePreciseDistanceAlongTrip(ctx context.Context, stopID string, shapePoints []gtfs.ShapePoint) float64 {
	if len(shapePoints) == 0 {
		return 0.0
	}

	// Get stop coordinates
	stop, err := api.GtfsManager.GtfsDB.Queries.GetStop(ctx, stopID)
	if err != nil {
		return 0.0
	}

	// Pre-calculate cumulative distances (this is inefficient for multiple stops)
	cumulativeDistances := preCalculateCumulativeDistances(shapePoints)

	return api.calculatePreciseDistanceAlongTripWithCoords(stop.Lat, stop.Lon, shapePoints, cumulativeDistances)
}

// preCalculateCumulativeDistances pre-calculates cumulative distances along shape points
// Returns an array where cumulativeDistances[i] is the cumulative distance up to (but not including) segment i
func preCalculateCumulativeDistances(shapePoints []gtfs.ShapePoint) []float64 {
	if len(shapePoints) <= 1 {
		return []float64{0}
	}

	cumulativeDistances := make([]float64, len(shapePoints))
	cumulativeDistances[0] = 0

	for i := 1; i < len(shapePoints); i++ {
		segmentDistance := utils.Distance(
			shapePoints[i-1].Latitude, shapePoints[i-1].Longitude,
			shapePoints[i].Latitude, shapePoints[i].Longitude,
		)
		cumulativeDistances[i] = cumulativeDistances[i-1] + segmentDistance
	}

	return cumulativeDistances
}

// Helper function to calculate distance from point to line segment
func distanceToLineSegment(px, py, x1, y1, x2, y2 float64) (distance, ratio float64) {
	dx := x2 - x1
	dy := y2 - y1

	if dx == 0 && dy == 0 {
		// Line segment is a point
		return utils.Distance(px, py, x1, y1), 0
	}

	// Calculate the parameter t for the projection of point onto the line
	t := ((px-x1)*dx + (py-y1)*dy) / (dx*dx + dy*dy)

	// Clamp t to [0, 1] to stay within the line segment
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	// Find the closest point on the line segment
	closestX := x1 + t*dx
	closestY := y1 + t*dy

	return utils.Distance(px, py, closestX, closestY), t
}

// IMPORTANT: Caller must hold manager.RLock() before calling this method.
func (api *RestAPI) GetSituationIDsForTrip(ctx context.Context, tripID string) []string {
	var routeID string
	var agencyID string

	if api.GtfsManager.GtfsDB != nil {
		trip, err := api.GtfsManager.GtfsDB.Queries.GetTrip(ctx, tripID)
		if err == nil {
			routeID = trip.RouteID
			route, err := api.GtfsManager.GtfsDB.Queries.GetRoute(ctx, routeID)
			if err == nil {
				agencyID = route.AgencyID
			}
		}
	}

	alerts := api.GtfsManager.GetAlertsByIDs(tripID, routeID, agencyID)

	situationIDs := []string{}
	for _, alert := range alerts {
		if alert.ID == "" {
			continue
		}
		if agencyID != "" {
			situationIDs = append(situationIDs, utils.FormCombinedID(agencyID, alert.ID))
		} else {
			situationIDs = append(situationIDs, alert.ID)
		}
	}

	return situationIDs
}

type TripAgencyResolver struct {
	RouteToAgency map[string]string
	TripToRoute   map[string]string
}

// NewTripAgencyResolver creates a new TripAgencyResolver that maps trip IDs to their respective agency IDs.
func NewTripAgencyResolver(allRoutes []gtfsdb.Route, allTrips []gtfsdb.Trip) *TripAgencyResolver {
	routeToAgency := make(map[string]string, len(allRoutes))
	for _, route := range allRoutes {
		routeToAgency[route.ID] = route.AgencyID
	}
	tripToRoute := make(map[string]string, len(allTrips))
	for _, trip := range allTrips {
		tripToRoute[trip.ID] = trip.RouteID
	}
	return &TripAgencyResolver{
		RouteToAgency: routeToAgency,
		TripToRoute:   tripToRoute,
	}
}

// GetAgencyNameByTripID retrieves the agency name for a given trip ID.
func (r *TripAgencyResolver) GetAgencyNameByTripID(tripID string) string {
	routeID := r.TripToRoute[tripID]

	agency := r.RouteToAgency[routeID]

	return agency
}

func (api *RestAPI) interpolateDistanceAtScheduledTime(
	scheduledTime int64,
	stopTimes []gtfsdb.StopTime,
	cumulativeDistances []float64,
) float64 {
	if len(stopTimes) == 0 {
		return 0
	}

	for i := 0; i < len(stopTimes)-1; i++ {
		fromStop := stopTimes[i]
		toStop := stopTimes[i+1]

		fromTime := fromStop.DepartureTime / 1e9
		toTime := toStop.ArrivalTime / 1e9

		if scheduledTime >= fromTime && scheduledTime <= toTime {
			if toTime == fromTime {
				return cumulativeDistances[i]
			}

			timeRatio := float64(scheduledTime-fromTime) / float64(toTime-fromTime)

			fromDistance := cumulativeDistances[i*len(cumulativeDistances)/len(stopTimes)]
			toDistance := cumulativeDistances[(i+1)*len(cumulativeDistances)/len(stopTimes)]

			return fromDistance + timeRatio*(toDistance-fromDistance)
		}
	}

	if scheduledTime < stopTimes[0].ArrivalTime/1e9 {
		return 0
	}

	return cumulativeDistances[len(cumulativeDistances)-1]
}

func (api *RestAPI) calculateOffsetForStop(
	stopID string,
	stopTimes []*gtfsdb.StopTime,
	currentTime time.Time,
	serviceDate time.Time,
	scheduleDeviation int,
) int {
	currentTimeSeconds := utils.CalculateSecondsSinceServiceDate(currentTime, serviceDate)

	for _, st := range stopTimes {
		if st.StopID == stopID {
			stopTimeSeconds := st.ArrivalTime / 1e9
			if stopTimeSeconds == 0 {
				stopTimeSeconds = st.DepartureTime / 1e9
			}
			predictedArrival := stopTimeSeconds + int64(scheduleDeviation)
			return int(predictedArrival - currentTimeSeconds)
		}
	}

	return 0
}

func (api *RestAPI) findNextStopAfter(
	currentStopID string,
	stopTimes []*gtfsdb.StopTime,
	currentTime time.Time,
	serviceDate time.Time,
	scheduleDeviation int,
) (stopID string, offset int) {
	if len(stopTimes) == 0 {
		return "", 0
	}

	currentTimeSeconds := utils.CalculateSecondsSinceServiceDate(currentTime, serviceDate)

	for i, st := range stopTimes {
		if st.StopID == currentStopID {
			if i+1 < len(stopTimes) {
				nextSt := stopTimes[i+1]
				stopTimeSeconds := nextSt.ArrivalTime / 1e9
				if stopTimeSeconds == 0 {
					stopTimeSeconds = nextSt.DepartureTime / 1e9
				}
				predictedArrival := stopTimeSeconds + int64(scheduleDeviation)
				return nextSt.StopID, int(predictedArrival - currentTimeSeconds)
			}
			break
		}
	}

	return "", 0
}

func (api *RestAPI) calculateBatchStopDistances(
	timeStops []gtfsdb.StopTime,
	shapePoints []gtfs.ShapePoint,
	stopCoords map[string]struct{ lat, lon float64 },
	agencyID string,
) []models.StopTime {

	stopTimesList := make([]models.StopTime, 0, len(timeStops))

	if len(shapePoints) < 2 {
		for _, stopTime := range timeStops {
			stopTimesList = append(stopTimesList, models.StopTime{
				StopID:              utils.FormCombinedID(agencyID, stopTime.StopID),
				ArrivalTime:         int(stopTime.ArrivalTime / 1e9),
				DepartureTime:       int(stopTime.DepartureTime / 1e9),
				StopHeadsign:        utils.NullStringOrEmpty(stopTime.StopHeadsign),
				DistanceAlongTrip:   0.0,
				HistoricalOccupancy: "",
			})
		}
		return stopTimesList
	}

	// Pre-calculate cumulative distances
	cumulativeDistances := preCalculateCumulativeDistances(shapePoints)
	if len(cumulativeDistances) != len(shapePoints) {
		for _, stopTime := range timeStops {
			stopTimesList = append(stopTimesList, models.StopTime{
				StopID:              utils.FormCombinedID(agencyID, stopTime.StopID),
				ArrivalTime:         int(stopTime.ArrivalTime / 1e9),
				DepartureTime:       int(stopTime.DepartureTime / 1e9),
				StopHeadsign:        utils.NullStringOrEmpty(stopTime.StopHeadsign),
				DistanceAlongTrip:   0.0,
				HistoricalOccupancy: "",
			})
		}
		return stopTimesList
	}

	lastMatchedIndex := 0

	for _, stopTime := range timeStops {
		var distanceAlongTrip float64

		// Only calculate if we have valid coordinates
		if coords, exists := stopCoords[stopTime.StopID]; exists {
			stopLat := coords.lat
			stopLon := coords.lon

			// ensure lastMatchedIndex didn't go out of bounds
			if lastMatchedIndex >= len(shapePoints)-1 {
				lastMatchedIndex = len(shapePoints) - 2
			}

			var minDistance = math.Inf(1)
			var closestSegmentIndex = lastMatchedIndex
			var projectionRatio float64

			// Early exit threshold to speed up search
			//This may be too conservative for some cases but helps performance significantly
			const earlyExitThresholdMeters = 100.0

			// Start from lastMatchedIndex
			for i := lastMatchedIndex; i < len(shapePoints)-1; i++ {
				distance, ratio := distanceToLineSegment(
					stopLat, stopLon,
					shapePoints[i].Latitude, shapePoints[i].Longitude,
					shapePoints[i+1].Latitude, shapePoints[i+1].Longitude,
				)

				if distance < minDistance {
					minDistance = distance
					closestSegmentIndex = i
					projectionRatio = ratio
					lastMatchedIndex = i
				} else if distance > minDistance+earlyExitThresholdMeters {
					// Early exit:
					break
				}
			}

			// Calculate distance along trip
			cumulativeDistance := cumulativeDistances[closestSegmentIndex]
			if closestSegmentIndex < len(shapePoints)-1 {
				segmentDistance := utils.Distance(
					shapePoints[closestSegmentIndex].Latitude, shapePoints[closestSegmentIndex].Longitude,
					shapePoints[closestSegmentIndex+1].Latitude, shapePoints[closestSegmentIndex+1].Longitude,
				)
				cumulativeDistance += segmentDistance * projectionRatio
			}
			distanceAlongTrip = cumulativeDistance
		}

		stopTimesList = append(stopTimesList, models.StopTime{
			StopID:              utils.FormCombinedID(agencyID, stopTime.StopID),
			ArrivalTime:         int(stopTime.ArrivalTime / 1e9),
			DepartureTime:       int(stopTime.DepartureTime / 1e9),
			StopHeadsign:        utils.NullStringOrEmpty(stopTime.StopHeadsign),
			DistanceAlongTrip:   distanceAlongTrip,
			HistoricalOccupancy: "",
		})
	}
	return stopTimesList
}

func (api *RestAPI) findStopsByScheduleDeviation(
	stopTimes []*gtfsdb.StopTime,
	currentTime time.Time,
	serviceDate time.Time,
	scheduleDeviation int,
) (closestStopID string, closestOffset int, nextStopID string, nextOffset int) {
	if len(stopTimes) == 0 {
		return "", 0, "", 0
	}

	currentTimeSeconds := utils.CalculateSecondsSinceServiceDate(currentTime, serviceDate)
	effectiveScheduleTime := currentTimeSeconds - int64(scheduleDeviation)

	var closestStop *gtfsdb.StopTime
	var closestTimeDiff int64 = math.MaxInt64

	for _, st := range stopTimes {
		stopTime := st.ArrivalTime / 1e9
		if stopTime == 0 {
			stopTime = st.DepartureTime / 1e9
		}

		timeDiff := stopTime - effectiveScheduleTime
		if timeDiff < 0 {
			timeDiff = -timeDiff
		}

		if timeDiff < closestTimeDiff {
			closestTimeDiff = timeDiff
			closestStop = st
		}
	}

	if closestStop == nil {
		return "", 0, "", 0
	}

	closestStopID = closestStop.StopID

	closestStopTime := closestStop.ArrivalTime / 1e9
	if closestStopTime == 0 {
		closestStopTime = closestStop.DepartureTime / 1e9
	}
	predictedClosestArrival := closestStopTime + int64(scheduleDeviation)
	closestOffset = int(predictedClosestArrival - currentTimeSeconds)

	for i, st := range stopTimes {
		if st.StopID == closestStopID {
			if i+1 < len(stopTimes) {
				nextSt := stopTimes[i+1]
				nextStopID = nextSt.StopID

				nextStopTime := nextSt.ArrivalTime / 1e9
				if nextStopTime == 0 {
					nextStopTime = nextSt.DepartureTime / 1e9
				}
				predictedNextArrival := nextStopTime + int64(scheduleDeviation)
				nextOffset = int(predictedNextArrival - currentTimeSeconds)
			}
			break
		}
	}

	return closestStopID, closestOffset, nextStopID, nextOffset
}

func (api *RestAPI) findClosestStopBySequence(
	stopTimes []*gtfsdb.StopTime,
	currentStopSequence uint32,
	currentTime time.Time,
	serviceDate time.Time,
	scheduleDeviation int,
	vehicle *gtfs.Vehicle,
) (stopID string, offset int) {
	currentTimeSeconds := utils.CalculateSecondsSinceServiceDate(currentTime, serviceDate)

	for _, st := range stopTimes {
		if uint32(st.StopSequence) == currentStopSequence {
			stopTimeSeconds := st.ArrivalTime / 1e9
			if stopTimeSeconds == 0 {
				stopTimeSeconds = st.DepartureTime / 1e9
			}
			predictedArrival := stopTimeSeconds + int64(scheduleDeviation)
			return st.StopID, int(predictedArrival - currentTimeSeconds)
		}
	}

	return "", 0
}

func (api *RestAPI) findNextStopBySequence(
	ctx context.Context,
	stopTimes []*gtfsdb.StopTime,
	currentStopSequence uint32,
	currentTime time.Time,
	serviceDate time.Time,
	scheduleDeviation int,
	vehicle *gtfs.Vehicle,
	tripID string,
	serviceDateForBlock time.Time,
) (stopID string, offset int) {
	currentTimeSeconds := utils.CalculateSecondsSinceServiceDate(currentTime, serviceDate)

	isAtCurrentStop := vehicle != nil && vehicle.CurrentStatus != nil &&
		*vehicle.CurrentStatus == gtfs.CurrentStatus(1)

	for i, st := range stopTimes {
		if uint32(st.StopSequence) == currentStopSequence {
			var nextStop *gtfsdb.StopTime

			if isAtCurrentStop {
				if i+1 < len(stopTimes) {
					nextStop = stopTimes[i+1]
				} else {
					nextStop = api.getFirstStopOfNextTripInBlock(ctx, tripID, serviceDateForBlock)
				}
			} else {
				nextStop = st
			}

			if nextStop != nil {
				stopTimeSeconds := nextStop.ArrivalTime / 1e9
				if stopTimeSeconds == 0 {
					stopTimeSeconds = nextStop.DepartureTime / 1e9
				}
				predictedArrival := stopTimeSeconds + int64(scheduleDeviation)
				return nextStop.StopID, int(predictedArrival - currentTimeSeconds)
			}
		}
	}

	return "", 0
}

func (api *RestAPI) getFirstStopOfNextTripInBlock(ctx context.Context, currentTripID string, serviceDate time.Time) *gtfsdb.StopTime {
	trip, err := api.GtfsManager.GtfsDB.Queries.GetTrip(ctx, currentTripID)
	if err != nil || !trip.BlockID.Valid {
		return nil
	}

	orderedTrips, err := api.GtfsManager.GtfsDB.Queries.GetTripsByBlockIDOrdered(ctx, gtfsdb.GetTripsByBlockIDOrderedParams{
		BlockID:    trip.BlockID,
		ServiceIds: []string{trip.ServiceID},
	})
	if err != nil {
		return nil
	}

	currentIndex := -1
	for i, t := range orderedTrips {
		if t.ID == currentTripID {
			currentIndex = i
			break
		}
	}

	if currentIndex >= 0 && currentIndex+1 < len(orderedTrips) {
		nextTripID := orderedTrips[currentIndex+1].ID
		nextTripStopTimes, err := api.GtfsManager.GtfsDB.Queries.GetStopTimesForTrip(ctx, nextTripID)
		if err == nil && len(nextTripStopTimes) > 0 {
			return &nextTripStopTimes[0]
		}
	}

	return nil
}

func (api *RestAPI) calculateEffectiveDistanceAlongTrip(
	ctx context.Context,
	actualDistance float64,
	scheduleDeviation int,
	currentTime time.Time,
	serviceDate time.Time,
	stopTimes []gtfsdb.StopTime,
	shapePoints []gtfs.ShapePoint,
	cumulativeDistances []float64,
) float64 {
	if scheduleDeviation == 0 || len(stopTimes) == 0 {
		return actualDistance
	}

	stopDistances := make([]float64, len(stopTimes))
	for i, st := range stopTimes {
		stop, err := api.GtfsManager.GtfsDB.Queries.GetStop(ctx, st.StopID)
		if err == nil {
			stopDistances[i] = api.calculatePreciseDistanceAlongTripWithCoords(
				stop.Lat, stop.Lon, shapePoints, cumulativeDistances,
			)
		}
	}

	currentTimeSeconds := utils.CalculateSecondsSinceServiceDate(currentTime, serviceDate)
	effectiveScheduleTime := currentTimeSeconds - int64(scheduleDeviation)

	return interpolateDistanceAtScheduledTime(effectiveScheduleTime, stopTimes, stopDistances)
}

func interpolateDistanceAtScheduledTime(
	scheduledTime int64,
	stopTimes []gtfsdb.StopTime,
	cumulativeDistances []float64,
) float64 {
	if len(stopTimes) == 0 {
		return 0
	}

	for i := 0; i < len(stopTimes)-1; i++ {
		fromStop := stopTimes[i]
		toStop := stopTimes[i+1]

		fromTime := fromStop.DepartureTime / 1e9
		toTime := toStop.ArrivalTime / 1e9

		if scheduledTime >= fromTime && scheduledTime <= toTime {
			if toTime == fromTime {
				return cumulativeDistances[i]
			}

			timeRatio := float64(scheduledTime-fromTime) / float64(toTime-fromTime)

			fromDistance := cumulativeDistances[i*len(cumulativeDistances)/len(stopTimes)]
			toDistance := cumulativeDistances[(i+1)*len(cumulativeDistances)/len(stopTimes)]

			return fromDistance + timeRatio*(toDistance-fromDistance)
		}
	}

	if scheduledTime < stopTimes[0].ArrivalTime/1e9 {
		return 0
	}

	return cumulativeDistances[len(cumulativeDistances)-1]
}
