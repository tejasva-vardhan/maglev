package restapi

import (
	"testing"
	"time"

	"github.com/OneBusAway/go-gtfs"
	"github.com/stretchr/testify/assert"
)

func TestGetScheduleDeviation_NoUpdates(t *testing.T) {
	api := createTestApi(t)
	defer api.Shutdown()

	deviation, hasData := api.GetScheduleDeviation("no-such-trip")
	assert.Equal(t, 0, deviation)
	assert.False(t, hasData, "no trip updates should return hasData=false")
}

func TestGetScheduleDeviation_TripLevelDelay(t *testing.T) {
	api := createTestApi(t)
	defer api.Shutdown()

	delay := 90 * time.Second
	api.GtfsManager.MockAddTripUpdate("trip-delay-test", &delay, nil)

	deviation, hasData := api.GetScheduleDeviation("trip-delay-test")
	assert.Equal(t, 90, deviation)
	assert.True(t, hasData)
}

func TestGetScheduleDeviation_StopLevelArrivalDelay(t *testing.T) {
	api := createTestApi(t)
	defer api.Shutdown()

	stopID := "stop-1"
	arrivalDelay := 60 * time.Second
	updates := []gtfs.StopTimeUpdate{
		{
			StopID:  &stopID,
			Arrival: &gtfs.StopTimeEvent{Delay: &arrivalDelay},
		},
	}
	api.GtfsManager.MockAddTripUpdate("trip-arrival-test", nil, updates)

	deviation, hasData := api.GetScheduleDeviation("trip-arrival-test")
	assert.Equal(t, 60, deviation)
	assert.True(t, hasData)
}

func TestGetScheduleDeviation_StopLevelDepartureDelay(t *testing.T) {
	api := createTestApi(t)
	defer api.Shutdown()

	stopID := "stop-1"
	departureDelay := 120 * time.Second
	updates := []gtfs.StopTimeUpdate{
		{
			StopID:    &stopID,
			Departure: &gtfs.StopTimeEvent{Delay: &departureDelay},
		},
	}
	api.GtfsManager.MockAddTripUpdate("trip-departure-test", nil, updates)

	deviation, hasData := api.GetScheduleDeviation("trip-departure-test")
	assert.Equal(t, 120, deviation)
	assert.True(t, hasData)
}

func TestGetScheduleDeviation_TripLevelDelayTakesPrecedence(t *testing.T) {
	api := createTestApi(t)
	defer api.Shutdown()

	tripDelay := 30 * time.Second
	stopID := "stop-1"
	stopDelay := 90 * time.Second
	updates := []gtfs.StopTimeUpdate{
		{
			StopID:  &stopID,
			Arrival: &gtfs.StopTimeEvent{Delay: &stopDelay},
		},
	}
	api.GtfsManager.MockAddTripUpdate("trip-precedence-test", &tripDelay, updates)

	deviation, hasData := api.GetScheduleDeviation("trip-precedence-test")
	assert.Equal(t, 30, deviation, "trip-level delay should take precedence over stop-level delay")
	assert.True(t, hasData)
}

func TestGetScheduleDeviation_StopUpdateWithNoDelay(t *testing.T) {
	api := createTestApi(t)
	defer api.Shutdown()

	stopID := "stop-1"
	updates := []gtfs.StopTimeUpdate{
		{
			StopID:  &stopID,
			Arrival: &gtfs.StopTimeEvent{}, // no Delay set
		},
	}
	api.GtfsManager.MockAddTripUpdate("trip-nodelay-test", nil, updates)

	deviation, hasData := api.GetScheduleDeviation("trip-nodelay-test")
	assert.Equal(t, 0, deviation)
	assert.False(t, hasData, "trip update with no delay data should report hasData=false")
}

func TestGetScheduleDeviation_ZeroDeviationIsDistinguishedFromNoData(t *testing.T) {
	api := createTestApi(t)
	defer api.Shutdown()

	// Trip with explicit zero-second delay — should return (0, true)
	zeroDelay := time.Duration(0)
	api.GtfsManager.MockAddTripUpdate("trip-zero-delay", &zeroDelay, nil)

	deviation, hasData := api.GetScheduleDeviation("trip-zero-delay")
	assert.Equal(t, 0, deviation)
	assert.True(t, hasData, "zero delay with trip update should still report hasData=true")

	// Nonexistent trip — should return (0, false)
	deviation2, hasData2 := api.GetScheduleDeviation("nonexistent-trip")
	assert.Equal(t, 0, deviation2)
	assert.False(t, hasData2, "nonexistent trip should report hasData=false")
}

func TestGetStopDelaysFromTripUpdates_NoUpdates(t *testing.T) {
	api := createTestApi(t)
	defer api.Shutdown()

	delays := api.GetStopDelaysFromTripUpdates("no-such-trip")
	assert.Empty(t, delays)
}

func TestGetStopDelaysFromTripUpdates_WithArrivalDelay(t *testing.T) {
	api := createTestApi(t)
	defer api.Shutdown()

	stopID := "stop-A"
	arrivalDelay := 45 * time.Second
	updates := []gtfs.StopTimeUpdate{
		{
			StopID:  &stopID,
			Arrival: &gtfs.StopTimeEvent{Delay: &arrivalDelay},
		},
	}
	api.GtfsManager.MockAddTripUpdate("trip-stop-delays-arrival", nil, updates)

	delays := api.GetStopDelaysFromTripUpdates("trip-stop-delays-arrival")
	assert.Len(t, delays, 1)
	assert.Equal(t, int64(45), delays["stop-A"].ArrivalDelay)
	assert.Equal(t, int64(0), delays["stop-A"].DepartureDelay)
}

func TestGetStopDelaysFromTripUpdates_WithDepartureDelay(t *testing.T) {
	api := createTestApi(t)
	defer api.Shutdown()

	stopID := "stop-B"
	departureDelay := 75 * time.Second
	updates := []gtfs.StopTimeUpdate{
		{
			StopID:    &stopID,
			Departure: &gtfs.StopTimeEvent{Delay: &departureDelay},
		},
	}
	api.GtfsManager.MockAddTripUpdate("trip-stop-delays-departure", nil, updates)

	delays := api.GetStopDelaysFromTripUpdates("trip-stop-delays-departure")
	assert.Len(t, delays, 1)
	assert.Equal(t, int64(0), delays["stop-B"].ArrivalDelay)
	assert.Equal(t, int64(75), delays["stop-B"].DepartureDelay)
}

func TestGetStopDelaysFromTripUpdates_SkipsStopWithNoStopID(t *testing.T) {
	api := createTestApi(t)
	defer api.Shutdown()

	arrivalDelay := 30 * time.Second
	updates := []gtfs.StopTimeUpdate{
		{
			StopID:  nil, // no stop ID — should be skipped
			Arrival: &gtfs.StopTimeEvent{Delay: &arrivalDelay},
		},
	}
	api.GtfsManager.MockAddTripUpdate("trip-nil-stopid", nil, updates)

	delays := api.GetStopDelaysFromTripUpdates("trip-nil-stopid")
	assert.Empty(t, delays, "stop updates without StopID should be skipped")
}

func TestGetStopDelaysFromTripUpdates_IncludesStopWithZeroDelays(t *testing.T) {
	api := createTestApi(t)
	defer api.Shutdown()

	stopID := "stop-C"
	zeroDelay := time.Duration(0)
	updates := []gtfs.StopTimeUpdate{
		{
			StopID:  &stopID,
			Arrival: &gtfs.StopTimeEvent{Delay: &zeroDelay},
		},
	}
	api.GtfsManager.MockAddTripUpdate("trip-zero-delays", nil, updates)

	delays := api.GetStopDelaysFromTripUpdates("trip-zero-delays")
	assert.Len(t, delays, 1, "stops with zero delays should be included")
	assert.Contains(t, delays, "stop-C")
	assert.Equal(t, int64(0), delays["stop-C"].ArrivalDelay)
}

func TestGetStopDelaysFromTripUpdates_MultipleStops(t *testing.T) {
	api := createTestApi(t)
	defer api.Shutdown()

	stopA := "stop-A"
	stopB := "stop-B"
	stopC := "stop-C"
	delayA := 30 * time.Second
	delayB := 60 * time.Second

	updates := []gtfs.StopTimeUpdate{
		{StopID: &stopA, Arrival: &gtfs.StopTimeEvent{Delay: &delayA}},
		{StopID: &stopB, Departure: &gtfs.StopTimeEvent{Delay: &delayB}},
		{StopID: &stopC}, // no delay events — still included with zero values
	}
	api.GtfsManager.MockAddTripUpdate("trip-multi-stops", nil, updates)

	delays := api.GetStopDelaysFromTripUpdates("trip-multi-stops")
	assert.Len(t, delays, 3, "all stops with StopID should be included")
	assert.Equal(t, int64(30), delays["stop-A"].ArrivalDelay)
	assert.Equal(t, int64(60), delays["stop-B"].DepartureDelay)
	assert.Contains(t, delays, "stop-C")
	assert.Equal(t, int64(0), delays["stop-C"].ArrivalDelay)
	assert.Equal(t, int64(0), delays["stop-C"].DepartureDelay)
}
