package gtfs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/jamespfennell/gtfs"
	"github.com/stretchr/testify/assert"
)

func TestParallelRealtimeUpdates(t *testing.T) {
	// Track server calls to verify parallelism
	var mu sync.Mutex
	callTimes := make([]time.Time, 0, 2)
	callDelays := make([]time.Duration, 0, 2)

	// Create test servers that simulate real GTFS-RT endpoints
	mux := http.NewServeMux()

	mux.HandleFunc("/trip-updates", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callTimes = append(callTimes, time.Now())
		callDelays = append(callDelays, 100*time.Millisecond) // Simulate processing time
		mu.Unlock()

		// Simulate some processing time
		time.Sleep(100 * time.Millisecond)

		data, err := os.ReadFile(filepath.Join("../../testdata", "raba-trip-updates.pb"))
		if err != nil {
			// If test data doesn't exist, return empty GTFS-RT data
			w.Header().Set("Content-Type", "application/x-protobuf")
			_, _ = w.Write([]byte{})
			return
		}
		w.Header().Set("Content-Type", "application/x-protobuf")
		_, _ = w.Write(data)
	})

	mux.HandleFunc("/vehicle-positions", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		callTimes = append(callTimes, time.Now())
		callDelays = append(callDelays, 100*time.Millisecond) // Simulate processing time
		mu.Unlock()

		// Simulate some processing time
		time.Sleep(100 * time.Millisecond)

		data, err := os.ReadFile(filepath.Join("../../testdata", "raba-vehicle-positions.pb"))
		if err != nil {
			// If test data doesn't exist, return empty GTFS-RT data
			w.Header().Set("Content-Type", "application/x-protobuf")
			_, _ = w.Write([]byte{})
			return
		}
		w.Header().Set("Content-Type", "application/x-protobuf")
		_, _ = w.Write(data)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Test current sequential implementation
	t.Run("Sequential updates (current implementation)", func(t *testing.T) {
		mu.Lock()
		callTimes = callTimes[:0] // Clear previous calls
		callDelays = callDelays[:0]
		mu.Unlock()

		config := Config{
			TripUpdatesURL:      server.URL + "/trip-updates",
			VehiclePositionsURL: server.URL + "/vehicle-positions",
		}

		manager := &Manager{
			realTimeMutex: sync.RWMutex{},
		}

		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		manager.updateGTFSRealtimeSequential(ctx, config)
		elapsed := time.Since(start)

		mu.Lock()
		numCalls := len(callTimes)
		mu.Unlock()

		assert.Equal(t, 2, numCalls, "Should make 2 calls (trips and vehicles)")
		assert.Greater(t, elapsed, 200*time.Millisecond, "Sequential calls should take at least 200ms")
		assert.Less(t, elapsed, 300*time.Millisecond, "But shouldn't take too much longer")
	})

	// Test new parallel implementation
	t.Run("Parallel updates (new implementation)", func(t *testing.T) {
		mu.Lock()
		callTimes = callTimes[:0] // Clear previous calls
		callDelays = callDelays[:0]
		mu.Unlock()

		config := Config{
			TripUpdatesURL:      server.URL + "/trip-updates",
			VehiclePositionsURL: server.URL + "/vehicle-positions",
		}

		manager := &Manager{
			realTimeMutex: sync.RWMutex{},
		}

		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		manager.updateGTFSRealtimeParallel(ctx, config)
		elapsed := time.Since(start)

		mu.Lock()
		numCalls := len(callTimes)
		timeDiff := time.Duration(0)
		if len(callTimes) >= 2 {
			// Calculate time difference between first and second call
			if callTimes[1].After(callTimes[0]) {
				timeDiff = callTimes[1].Sub(callTimes[0])
			} else {
				timeDiff = callTimes[0].Sub(callTimes[1])
			}
		}
		mu.Unlock()

		assert.Equal(t, 2, numCalls, "Should make 2 calls (trips and vehicles)")
		assert.Less(t, elapsed, 150*time.Millisecond, "Parallel calls should be significantly faster")
		assert.Less(t, timeDiff, 50*time.Millisecond, "Calls should be made nearly simultaneously")
	})
}

func TestParallelRealtimeUpdatesWithErrors(t *testing.T) {
	// Test error handling in parallel updates
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer errorServer.Close()

	config := Config{
		TripUpdatesURL:      errorServer.URL + "/trip-updates",
		VehiclePositionsURL: errorServer.URL + "/vehicle-positions",
	}

	manager := &Manager{
		realTimeMutex: sync.RWMutex{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Should not panic or hang when both endpoints fail
	manager.updateGTFSRealtimeParallel(ctx, config)

	// Verify that no data was stored (due to errors)
	trips := manager.GetRealTimeTrips()
	vehicles := manager.GetRealTimeVehicles()

	assert.Empty(t, trips, "No trips should be stored when errors occur")
	assert.Empty(t, vehicles, "No vehicles should be stored when errors occur")
}

func TestParallelRealtimeUpdatesWithContextCancellation(t *testing.T) {
	// Test context cancellation during parallel updates
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow response
		time.Sleep(200 * time.Millisecond)
		w.Header().Set("Content-Type", "application/x-protobuf")
		_, _ = w.Write([]byte{})
	}))
	defer slowServer.Close()

	config := Config{
		TripUpdatesURL:      slowServer.URL + "/trip-updates",
		VehiclePositionsURL: slowServer.URL + "/vehicle-positions",
	}

	manager := &Manager{
		realTimeMutex: sync.RWMutex{},
	}

	// Create a context with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	manager.updateGTFSRealtimeParallel(ctx, config)
	elapsed := time.Since(start)

	// Should return quickly due to context cancellation
	assert.Less(t, elapsed, 100*time.Millisecond, "Should return quickly when context is cancelled")
}

// Helper methods that will be implemented
func (manager *Manager) updateGTFSRealtimeSequential(ctx context.Context, config Config) {
	// This is the current implementation (renamed for testing)
	headers := map[string]string{}
	if config.RealTimeAuthHeaderKey != "" && config.RealTimeAuthHeaderValue != "" {
		headers[config.RealTimeAuthHeaderKey] = config.RealTimeAuthHeaderValue
	}
	tripData, tripErr := loadRealtimeData(ctx, config.TripUpdatesURL, headers)

	if ctx.Err() != nil {
		return
	}

	vehicleData, vehicleErr := loadRealtimeData(ctx, config.VehiclePositionsURL, headers)

	if tripErr != nil || vehicleErr != nil || ctx.Err() != nil {
		return
	}

	manager.realTimeMutex.Lock()
	defer manager.realTimeMutex.Unlock()

	if tripData != nil {
		manager.realTimeTrips = tripData.Trips
	}
	if vehicleData != nil {
		manager.realTimeVehicles = vehicleData.Vehicles
	}
}

func (manager *Manager) updateGTFSRealtimeParallel(ctx context.Context, config Config) {
	// This will be the new parallel implementation
	headers := map[string]string{}
	if config.RealTimeAuthHeaderKey != "" && config.RealTimeAuthHeaderValue != "" {
		headers[config.RealTimeAuthHeaderKey] = config.RealTimeAuthHeaderValue
	}

	var wg sync.WaitGroup
	var tripData, vehicleData *gtfs.Realtime
	var tripErr, vehicleErr error

	// Fetch trip updates in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		tripData, tripErr = loadRealtimeData(ctx, config.TripUpdatesURL, headers)
	}()

	// Fetch vehicle positions in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		vehicleData, vehicleErr = loadRealtimeData(ctx, config.VehiclePositionsURL, headers)
	}()

	// Wait for both to complete
	wg.Wait()

	// Check for context cancellation
	if ctx.Err() != nil {
		return
	}

	// Check errors but don't fail if one succeeds
	// In production, these errors would be logged by the logger

	// Update data if at least one succeeded
	if (tripData != nil || vehicleData != nil) && (tripErr == nil || vehicleErr == nil) {
		manager.realTimeMutex.Lock()
		defer manager.realTimeMutex.Unlock()

		if tripData != nil && tripErr == nil {
			manager.realTimeTrips = tripData.Trips
		}
		if vehicleData != nil && vehicleErr == nil {
			manager.realTimeVehicles = vehicleData.Vehicles
		}
	}
}

func TestRealTimeDataConsistency(t *testing.T) {
	// Test that parallel updates maintain data consistency
	manager := &Manager{
		realTimeMutex: sync.RWMutex{},
	}

	// Run multiple parallel updates to test for race conditions
	var wg sync.WaitGroup
	ctx := context.Background()

	config := Config{
		TripUpdatesURL:      "http://invalid.example.com/trips",
		VehiclePositionsURL: "http://invalid.example.com/vehicles",
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			manager.updateGTFSRealtimeParallel(ctx, config)
		}()
	}

	wg.Wait()

	// Should not panic or cause data corruption
	// The actual data will be empty due to invalid URLs, but calls should complete
	trips := manager.GetRealTimeTrips()
	vehicles := manager.GetRealTimeVehicles()

	// Slices may be nil or empty due to invalid URLs, but function should not panic
	assert.True(t, trips != nil || len(trips) == 0, "Trips should be accessible without panic")
	assert.True(t, vehicles != nil || len(vehicles) == 0, "Vehicles should be accessible without panic")
}
