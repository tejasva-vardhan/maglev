package restapi

import (
	"context"
	"math"
	"time"

	"github.com/OneBusAway/go-gtfs"
	gtfsrt "github.com/OneBusAway/go-gtfs/proto"
	"maglev.onebusaway.org/internal/models"
	"maglev.onebusaway.org/internal/utils"
)

// StaleDetector checks whether a vehicle's real-time data is too old to trust.
//
// Java reference: GtfsRealtimeSource.handleCombinedUpdates()
// onebusaway-transit-data-federation/src/main/java/org/onebusaway/transit_data_federation/impl/realtime/gtfs_realtime/GtfsRealtimeSource.java
//
// The Java implementation removes vehicle records from the active map when their
// last update time is more than 15 minutes in the past. We mirror that threshold
// here so stale vehicles are treated as absent rather than as live vehicles.
type StaleDetector struct {
	threshold time.Duration
}

func NewStaleDetector() *StaleDetector {
	return &StaleDetector{threshold: 15 * time.Minute}
}

func (d *StaleDetector) WithThreshold(threshold time.Duration) *StaleDetector {
	d.threshold = threshold
	return d
}

// Check returns true when the vehicle's timestamp is missing or older than the threshold.
func (d *StaleDetector) Check(vehicle *gtfs.Vehicle, currentTime time.Time) bool {
	if vehicle == nil || vehicle.Timestamp == nil {
		return true
	}
	return currentTime.Sub(*vehicle.Timestamp) > d.threshold
}

var defaultStaleDetector = NewStaleDetector()

// scheduleRelationshipStatus converts a GTFS-RT TripDescriptor_ScheduleRelationship to
// the OBA status string.
//
// Java reference: GtfsRealtimeTripLibrary.java
// onebusaway-transit-data-federation/src/main/java/org/onebusaway/transit_data_federation/impl/realtime/gtfs_realtime/GtfsRealtimeTripLibrary.java
//
// Java calls record.setStatus(blockDescriptor.getScheduleRelationship().toString()), which
// produces the enum name as-is ("SCHEDULED", "CANCELED", "ADDED", "DUPLICATED"). The status
// "default" is only used when no real-time data exists at all (TripStatusBeanServiceImpl line 253).
func scheduleRelationshipStatus(sr gtfs.TripScheduleRelationship) string {
	switch sr {
	case gtfsrt.TripDescriptor_CANCELED:
		return "CANCELED"
	case gtfsrt.TripDescriptor_ADDED:
		return "ADDED"
	case gtfsrt.TripDescriptor_DUPLICATED:
		return "DUPLICATED"
	default:
		return "SCHEDULED"
	}
}

/*
Note!!
GetVehicleStatusAndPhase returns the OBA status and phase for a vehicle.
Java reference: VehicleStatusServiceImpl.java (handleVehicleLocationRecord)
onebusaway-transit-data-federation/src/main/java/org/onebusaway/transit_data_federation/impl/realtime/VehicleStatusServiceImpl.java
The Java implementation does not map directly to GTFS-RT CurrentStatus values.
Instead, it uses a simple rule: if a vehicle location record has been received,
the trip is "in_progress"; otherwise it remains "scheduled". The phase is
determined solely by the presence of the vehicle, not by its GTFS-RT stop status.
Status comes from the trip's schedule relationship ("SCHEDULED", "CANCELED", "ADDED", "DUPLICATED").
"default" is only returned when no real-time data exists at all.
*/
func GetVehicleStatusAndPhase(vehicle *gtfs.Vehicle) (status string, phase string) {
	if vehicle == nil {
		return "default", ""
	}

	sr := gtfsrt.TripDescriptor_SCHEDULED
	if vehicle.Trip != nil {
		sr = vehicle.Trip.ID.ScheduleRelationship
	}
	status = scheduleRelationshipStatus(sr)

	if vehicle.CurrentStatus != nil {
		phase = "in_progress"
	}

	return status, phase
}

func (api *RestAPI) BuildVehicleStatus(
	ctx context.Context,
	vehicle *gtfs.Vehicle,
	tripID string,
	agencyID string,
	status *models.TripStatusForTripDetails,
) {
	if vehicle == nil || defaultStaleDetector.Check(vehicle, time.Now()) {
		status.Status, status.Phase = GetVehicleStatusAndPhase(nil)
		return
	}

	if vehicle.Timestamp != nil {
		status.LastUpdateTime = api.GtfsManager.GetVehicleLastUpdateTime(vehicle)
	}

	if vehicle.Position != nil && vehicle.Position.Latitude != nil && vehicle.Position.Longitude != nil {
		actualPosition := models.Location{
			Lat: float64(*vehicle.Position.Latitude),
			Lon: float64(*vehicle.Position.Longitude),
		}
		status.LastKnownLocation = actualPosition

		projectedPosition := api.projectPositionOntoRoute(ctx, tripID, actualPosition)
		if projectedPosition != nil {
			status.Position = *projectedPosition
		} else {
			status.Position = actualPosition
		}

		if vehicle.Timestamp != nil {
			status.LastLocationUpdateTime = api.GtfsManager.GetVehicleLastUpdateTime(vehicle)
		}
	}

	if vehicle.Position != nil && vehicle.Position.Bearing != nil {
		obaOrientation := (90 - *vehicle.Position.Bearing)
		if obaOrientation < 0 {
			obaOrientation += 360
		}
		status.Orientation = float64(obaOrientation)
		status.LastKnownOrientation = float64(obaOrientation)
	}

	status.Status, status.Phase = GetVehicleStatusAndPhase(vehicle)

	if vehicle.Trip != nil && vehicle.Trip.ID.ID != "" {
		status.ActiveTripID = utils.FormCombinedID(agencyID, vehicle.Trip.ID.ID)
	} else {
		status.ActiveTripID = utils.FormCombinedID(agencyID, tripID)
	}

	hasRealtimeData := vehicle.Position != nil || vehicle.Timestamp != nil
	status.Predicted = hasRealtimeData
	status.Scheduled = !hasRealtimeData
}

func GetVehicleActiveTripID(vehicle *gtfs.Vehicle) string {
	if vehicle == nil || vehicle.Trip == nil || vehicle.Trip.ID.ID == "" {
		return ""
	}

	return vehicle.Trip.ID.ID
}

func (api *RestAPI) projectPositionOntoRoute(ctx context.Context, tripID string, actualPos models.Location) *models.Location {
	shapeRows, err := api.GtfsManager.GtfsDB.Queries.GetShapePointsByTripID(ctx, tripID)
	if err != nil || len(shapeRows) < 2 {
		return nil
	}

	shapePoints := make([]gtfs.ShapePoint, len(shapeRows))
	for i, sp := range shapeRows {
		shapePoints[i] = gtfs.ShapePoint{
			Latitude:  sp.Lat,
			Longitude: sp.Lon,
		}
	}

	minDistance := math.MaxFloat64
	var closestPoint models.Location

	for i := 0; i < len(shapePoints)-1; i++ {
		distance, projectedPoint := projectPointToSegment(
			actualPos.Lat, actualPos.Lon,
			shapePoints[i].Latitude, shapePoints[i].Longitude,
			shapePoints[i+1].Latitude, shapePoints[i+1].Longitude,
		)

		if distance < minDistance {
			minDistance = distance
			closestPoint = projectedPoint
		}
	}

	if minDistance <= 200 {
		return &closestPoint
	}

	return nil
}

func projectPointToSegment(px, py, x1, y1, x2, y2 float64) (float64, models.Location) {
	dx := x2 - x1
	dy := y2 - y1

	if dx == 0 && dy == 0 {
		dist := utils.Distance(px, py, x1, y1)
		return dist, models.Location{Lat: x1, Lon: y1}
	}

	t := ((px-x1)*dx + (py-y1)*dy) / (dx*dx + dy*dy)

	if t < 0 {
		dist := utils.Distance(px, py, x1, y1)
		return dist, models.Location{Lat: x1, Lon: y1}
	}
	if t > 1 {
		dist := utils.Distance(px, py, x2, y2)
		return dist, models.Location{Lat: x2, Lon: y2}
	}

	projLat := x1 + t*dx
	projLon := y1 + t*dy

	dist := utils.Distance(px, py, projLat, projLon)
	return dist, models.Location{Lat: projLat, Lon: projLon}
}

func getCurrentVehicleStopSequence(vehicle *gtfs.Vehicle) *int {
	if vehicle == nil || vehicle.CurrentStopSequence == nil {
		return nil
	}
	val := int(*vehicle.CurrentStopSequence)
	return &val
}
