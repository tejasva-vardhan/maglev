package gtfs

import (
	"fmt"
	"sync"
	"testing"

	"github.com/OneBusAway/go-gtfs"
)

// Benchmark for map rebuild optimization
func BenchmarkRebuildRealTimeTripLookup(b *testing.B) {
	manager := &Manager{
		realTimeMutex: sync.RWMutex{},
		feedTrips:     make(map[string][]gtfs.Trip),
	}

	feedTrips := make([]gtfs.Trip, 1000)
	for i := 0; i < 1000; i++ {
		feedTrips[i] = gtfs.Trip{
			ID: gtfs.TripID{ID: fmt.Sprintf("trip_%d", i)},
		}
	}
	manager.feedTrips["feed-0"] = feedTrips

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.rebuildMergedRealtimeLocked()
	}
}

func BenchmarkRebuildRealTimeVehicleLookupByTrip(b *testing.B) {
	manager := &Manager{
		realTimeMutex: sync.RWMutex{},
		feedVehicles:  make(map[string][]gtfs.Vehicle),
	}

	feedVehicles := make([]gtfs.Vehicle, 1000)
	for i := 0; i < 1000; i++ {
		feedVehicles[i] = gtfs.Vehicle{
			Trip: &gtfs.Trip{
				ID: gtfs.TripID{ID: fmt.Sprintf("trip_%d", i)},
			},
		}
	}
	manager.feedVehicles["feed-0"] = feedVehicles

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.rebuildMergedRealtimeLocked()
	}
}

func BenchmarkRebuildRealTimeVehicleLookupByVehicle(b *testing.B) {
	manager := &Manager{
		realTimeMutex: sync.RWMutex{},
		feedVehicles:  make(map[string][]gtfs.Vehicle),
	}

	feedVehicles := make([]gtfs.Vehicle, 1000)
	for i := 0; i < 1000; i++ {
		feedVehicles[i] = gtfs.Vehicle{
			ID: &gtfs.VehicleID{ID: fmt.Sprintf("vehicle_%d", i)},
		}
	}
	manager.feedVehicles["feed-0"] = feedVehicles

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.rebuildMergedRealtimeLocked()
	}
}
