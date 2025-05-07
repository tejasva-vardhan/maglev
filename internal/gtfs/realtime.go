package gtfs

import (
	"context"
	"github.com/jamespfennell/gtfs"
	"io"
	"log"
	"net/http"
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

func loadRealtimeData(ctx context.Context, source string) (*gtfs.Realtime, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", source, nil)
	if err != nil {
		return nil, err
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

func (manager *Manager) updateGTFSRealtime(ctx context.Context, tripUpdatesURL string, vehiclePositionsURL string) {
	tripData, tripErr := loadRealtimeData(ctx, tripUpdatesURL)

	if ctx.Err() != nil {
		return
	}

	vehicleData, vehicleErr := loadRealtimeData(ctx, vehiclePositionsURL)

	if tripErr != nil {
		log.Printf("Error loading GTFS-RT trip updates data from %s: %v", tripUpdatesURL, tripErr)
	}
	if vehicleErr != nil {
		log.Printf("Error loading GTFS-RT vehicle positions data from %s: %v", vehiclePositionsURL, vehicleErr)
	}

	if tripErr != nil || vehicleErr != nil || ctx.Err() != nil {
		return
	}

	manager.realTimeMutex.Lock()
	defer manager.realTimeMutex.Unlock()

	manager.realTimeTrips = tripData.Trips
	manager.realTimeVehicles = vehicleData.Vehicles
}

func (manager *Manager) updateGTFSRealtimePeriodically(tripUpdatesURL string, vehiclePositionsURL string) {
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
			manager.updateGTFSRealtime(ctx, tripUpdatesURL, vehiclePositionsURL)
			cancel() // Ensure the context is canceled when done
		}
	}
}
