package gtfs

import (
	"context"
	"github.com/jamespfennell/gtfs"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// GetRealTimeTrips returns the real-time trip updates
func (manager *Manager) GetRealTimeTrips() []gtfs.Trip {
	manager.realTimeMutex.RLock()
	defer manager.realTimeMutex.RUnlock()
	return manager.realTimeTrips
}

// GetRealTimeVehicles returns the real-time vehicle positions
func (manager *Manager) GetRealTimeVehicles() []gtfs.Vehicle {
	manager.realTimeMutex.RLock()
	defer manager.realTimeMutex.RUnlock()
	return manager.realTimeVehicles
}

func loadRealtimeData(ctx context.Context, source string, headers map[string]string) (*gtfs.Realtime, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", source, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // nolint: errcheck

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return gtfs.ParseRealtime(b, &gtfs.ParseRealtimeOptions{})
}

func (manager *Manager) updateGTFSRealtime(ctx context.Context, config Config) {
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
		if tripErr != nil {
			log.Printf("Error loading GTFS-RT trip updates data from %s: %v", config.TripUpdatesURL, tripErr)
		}
	}()

	// Fetch vehicle positions in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		vehicleData, vehicleErr = loadRealtimeData(ctx, config.VehiclePositionsURL, headers)
		if vehicleErr != nil {
			log.Printf("Error loading GTFS-RT vehicle positions data from %s: %v", config.VehiclePositionsURL, vehicleErr)
		}
	}()

	// Wait for both to complete
	wg.Wait()

	// Check for context cancellation
	if ctx.Err() != nil {
		return
	}

	// Update data if at least one fetch succeeded
	manager.realTimeMutex.Lock()
	defer manager.realTimeMutex.Unlock()

	if tripData != nil && tripErr == nil {
		manager.realTimeTrips = tripData.Trips
	}
	if vehicleData != nil && vehicleErr == nil {
		manager.realTimeVehicles = vehicleData.Vehicles
	}
}

func (manager *Manager) updateGTFSRealtimePeriodically(config Config) {
	defer manager.wg.Done()

	// Update every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for { // nolint
		select {
		case <-ticker.C:
			// Create a context with timeout for the download
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

			// Download realtime data
			log.Println("Updating GTFS-RT data")
			manager.updateGTFSRealtime(ctx, config)
			cancel() // Ensure the context is canceled when done
		case <-manager.shutdownChan:
			log.Println("Shutting down real-time updates")
			return
		}
	}
}
