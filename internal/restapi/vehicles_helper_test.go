package restapi

import (
	"context"
	"testing"
	"time"

	"github.com/OneBusAway/go-gtfs"
	gtfsrt "github.com/OneBusAway/go-gtfs/proto"
	"github.com/stretchr/testify/assert"
	"maglev.onebusaway.org/internal/models"
)

func TestGetVehicleStatusAndPhase_NilVehicle(t *testing.T) {
	status, phase := GetVehicleStatusAndPhase(nil)
	assert.Equal(t, "default", status)
	assert.Equal(t, "scheduled", phase)
}

func TestGetVehicleStatusAndPhase_ScheduledTrip(t *testing.T) {
	sr := gtfsrt.TripDescriptor_SCHEDULED
	vehicle := &gtfs.Vehicle{
		Trip: &gtfs.Trip{
			ID: gtfs.TripID{ScheduleRelationship: sr},
		},
	}
	status, phase := GetVehicleStatusAndPhase(vehicle)
	assert.Equal(t, "SCHEDULED", status)
	assert.Equal(t, "in_progress", phase)
}

func TestGetVehicleStatusAndPhase_CanceledTrip(t *testing.T) {
	sr := gtfsrt.TripDescriptor_CANCELED
	vehicle := &gtfs.Vehicle{
		Trip: &gtfs.Trip{
			ID: gtfs.TripID{ScheduleRelationship: sr},
		},
	}
	status, phase := GetVehicleStatusAndPhase(vehicle)
	assert.Equal(t, "CANCELED", status)
	assert.Equal(t, "", phase, "canceled trip should have empty phase")
}

func TestGetVehicleStatusAndPhase_AddedTrip(t *testing.T) {
	sr := gtfsrt.TripDescriptor_ADDED
	vehicle := &gtfs.Vehicle{
		Trip: &gtfs.Trip{
			ID: gtfs.TripID{ScheduleRelationship: sr},
		},
	}
	status, phase := GetVehicleStatusAndPhase(vehicle)
	assert.Equal(t, "ADDED", status)
	assert.Equal(t, "in_progress", phase)
}

func TestGetVehicleStatusAndPhase_DuplicatedTrip(t *testing.T) {
	sr := gtfsrt.TripDescriptor_DUPLICATED
	vehicle := &gtfs.Vehicle{
		Trip: &gtfs.Trip{
			ID: gtfs.TripID{ScheduleRelationship: sr},
		},
	}
	status, phase := GetVehicleStatusAndPhase(vehicle)
	assert.Equal(t, "DUPLICATED", status)
	assert.Equal(t, "in_progress", phase)
}

func TestGetVehicleStatusAndPhase_NoTripInfo(t *testing.T) {
	// Vehicle present but no Trip field — should default to SCHEDULED
	vehicle := &gtfs.Vehicle{}
	status, phase := GetVehicleStatusAndPhase(vehicle)
	assert.Equal(t, "SCHEDULED", status)
	assert.Equal(t, "in_progress", phase)
}

func TestStaleDetector_NilVehicle(t *testing.T) {
	d := NewStaleDetector()
	assert.True(t, d.Check(nil, time.Now()), "nil vehicle should be considered stale")
}

func TestStaleDetector_NilTimestamp(t *testing.T) {
	d := NewStaleDetector()
	vehicle := &gtfs.Vehicle{} // Timestamp is nil
	assert.True(t, d.Check(vehicle, time.Now()), "vehicle with nil timestamp should be considered stale")
}

func TestStaleDetector_FreshVehicle(t *testing.T) {
	d := NewStaleDetector()
	now := time.Now()
	recent := now.Add(-5 * time.Minute)
	vehicle := &gtfs.Vehicle{Timestamp: &recent}
	assert.False(t, d.Check(vehicle, now), "vehicle updated 5 minutes ago should not be stale with 15-minute threshold")
}

func TestStaleDetector_StaleVehicle(t *testing.T) {
	d := NewStaleDetector()
	now := time.Now()
	old := now.Add(-20 * time.Minute)
	vehicle := &gtfs.Vehicle{Timestamp: &old}
	assert.True(t, d.Check(vehicle, now), "vehicle updated 20 minutes ago should be stale with 15-minute threshold")
}

func TestStaleDetector_ExactThreshold(t *testing.T) {
	d := NewStaleDetector()
	now := time.Now()
	exactly := now.Add(-15 * time.Minute)
	vehicle := &gtfs.Vehicle{Timestamp: &exactly}
	// Exactly at threshold is NOT stale (threshold is strict >)
	assert.False(t, d.Check(vehicle, now), "vehicle at exactly 15 minutes should not be stale")
}

func TestStaleDetector_WithCustomThreshold(t *testing.T) {
	d := NewStaleDetector().WithThreshold(5 * time.Minute)
	now := time.Now()

	fresh := now.Add(-3 * time.Minute)
	freshVehicle := &gtfs.Vehicle{Timestamp: &fresh}
	assert.False(t, d.Check(freshVehicle, now), "3-minute old vehicle should not be stale with 5-minute threshold")

	stale := now.Add(-6 * time.Minute)
	staleVehicle := &gtfs.Vehicle{Timestamp: &stale}
	assert.True(t, d.Check(staleVehicle, now), "6-minute old vehicle should be stale with 5-minute threshold")
}

func TestBuildVehicleStatus_NilVehicleSetsDefaultStatus(t *testing.T) {
	api := createTestApi(t)
	defer api.Shutdown()
	ctx := context.Background()

	status := &models.TripStatusForTripDetails{}
	api.BuildVehicleStatus(ctx, nil, "any-trip", "any-agency", status)

	assert.Equal(t, "default", status.Status)
	assert.Equal(t, "scheduled", status.Phase)
	assert.False(t, status.Predicted, "BuildVehicleStatus must not set Predicted")
}

func TestBuildVehicleStatus_StaleVehicleSetsDefaultStatus(t *testing.T) {
	api := createTestApi(t)
	defer api.Shutdown()
	ctx := context.Background()

	old := time.Now().Add(-20 * time.Minute)
	vehicle := &gtfs.Vehicle{
		ID:        &gtfs.VehicleID{ID: "v1"},
		Timestamp: &old,
	}

	status := &models.TripStatusForTripDetails{}
	api.BuildVehicleStatus(ctx, vehicle, "any-trip", "any-agency", status)

	assert.Equal(t, "default", status.Status)
	assert.Equal(t, "scheduled", status.Phase)
}

func TestBuildVehicleStatus_FreshVehicleWithPosition_SetsLocationAndPhase(t *testing.T) {
	api := createTestApi(t)
	defer api.Shutdown()
	ctx := context.Background()

	now := time.Now()
	lat := float32(37.7749)
	lon := float32(-122.4194)
	vehicle := &gtfs.Vehicle{
		ID:        &gtfs.VehicleID{ID: "v1"},
		Timestamp: &now,
		Position: &gtfs.Position{
			Latitude:  &lat,
			Longitude: &lon,
		},
	}

	status := &models.TripStatusForTripDetails{}
	api.BuildVehicleStatus(ctx, vehicle, "any-trip", "any-agency", status)

	assert.False(t, status.Predicted, "BuildVehicleStatus must not set Predicted")
	assert.Equal(t, "SCHEDULED", status.Status)
	assert.Equal(t, "in_progress", status.Phase)
}

func TestBuildVehicleStatus_FreshVehicleNoPosition_DoesNotSetPredicted(t *testing.T) {
	api := createTestApi(t)
	defer api.Shutdown()
	ctx := context.Background()

	now := time.Now()
	vehicle := &gtfs.Vehicle{
		ID:        &gtfs.VehicleID{ID: "v1"},
		Timestamp: &now,
		// No Position
	}

	status := &models.TripStatusForTripDetails{}
	api.BuildVehicleStatus(ctx, vehicle, "any-trip", "any-agency", status)

	assert.False(t, status.Predicted, "BuildVehicleStatus must not set Predicted")
}
