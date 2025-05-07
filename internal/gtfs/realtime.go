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

func loadRealtimeData(source string) (*gtfs.Realtime, error) {
	resp, err := http.Get(source)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return gtfs.ParseRealtime(b, &gtfs.ParseRealtimeOptions{})
}

func (manager *Manager) updateGTFSRealtime(tripUpdatesURL string, vehiclePositionsURL string) {
	tripData, tripErr := loadRealtimeData(tripUpdatesURL)
	vehicleData, vehicleErr := loadRealtimeData(vehiclePositionsURL)

	if tripErr != nil || vehicleErr != nil {
		log.Printf("Error loading GTFS-RT data: %v, %v", tripErr, vehicleErr)
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
			_, cancel := context.WithTimeout(context.Background(), 15*time.Second)

			// Download realtime data
			log.Println("Updating GTFS-RT data")
			manager.updateGTFSRealtime(tripUpdatesURL, vehiclePositionsURL)
			cancel() // Always cancel the context when done
		}
	}
}
