package gtfs

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jamespfennell/gtfs"
)

// Manager manages the GTFS data and provides methods to access it
type Manager struct {
	gtfsSource       string
	gtfsData         *gtfs.Static
	lastUpdated      time.Time
	isLocalFile      bool
	realTimeTrips    []gtfs.Trip
	realTimeVehicles []gtfs.Vehicle
	realTimeMutex    sync.RWMutex
}

// InitGTFSManager initializes the Manager with the GTFS data from the given source
// The source can be either a URL or a local file path
func InitGTFSManager(config Config) (*Manager, error) {
	isLocalFile := !strings.HasPrefix(config.GtfsURL, "http://") && !strings.HasPrefix(config.GtfsURL, "https://")

	staticData, err := loadGTFSData(config.GtfsURL, isLocalFile)
	if err != nil {
		return nil, err
	}

	manager := &Manager{
		gtfsSource:  config.GtfsURL,
		gtfsData:    staticData,
		lastUpdated: time.Now(),
		isLocalFile: isLocalFile,
	}

	if !isLocalFile {
		go manager.updateStaticGTFS()
	}

	if config.realTimeDataEnabled() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel() // Ensure the context is canceled when done
		manager.updateGTFSRealtime(ctx, config)
		go manager.updateGTFSRealtimePeriodically(config)
	}

	return manager, nil
}

func (manager *Manager) GetRegionBounds() (lat, lon, latSpan, lonSpan float64) {
	var minLat, maxLat, minLon, maxLon float64
	first := true
	for _, shape := range manager.gtfsData.Shapes {
		for _, point := range shape.Points {
			if first {
				minLat = point.Latitude
				maxLat = point.Latitude
				minLon = point.Longitude
				maxLon = point.Longitude
				first = false
				continue
			}

			if point.Latitude < minLat {
				minLat = point.Latitude
			}
			if point.Latitude > maxLat {
				maxLat = point.Latitude
			}
			if point.Longitude < minLon {
				minLon = point.Longitude
			}
			if point.Longitude > maxLon {
				maxLon = point.Longitude
			}
		}
	}

	lat = (minLat + maxLat) / 2
	lon = (minLon + maxLon) / 2
	latSpan = maxLat - minLat
	lonSpan = maxLon - minLon

	return lat, lon, latSpan, lonSpan
}

func (manager *Manager) GetAgencies() []gtfs.Agency {
	return manager.gtfsData.Agencies
}

func (manager *Manager) GetStaticData() *gtfs.Static {
	return manager.gtfsData
}

func (manager *Manager) FindAgency(id string) *gtfs.Agency {
	for _, agency := range manager.gtfsData.Agencies {
		if agency.Id == id {
			return &agency
		}
	}
	return nil
}

// GetRoutesByAgencyID retrieves all routes associated with the specified agency ID from the GTFS data.
func (manager *Manager) GetRoutesByAgencyID(agencyID string) []*gtfs.Route {
	var agencyRoutes []*gtfs.Route

	for i := range manager.gtfsData.Routes {
		if manager.gtfsData.Routes[i].Agency.Id == agencyID {
			agencyRoutes = append(agencyRoutes, &manager.gtfsData.Routes[i])
		}
	}

	return agencyRoutes
}

func (manager *Manager) PrintStatistics() {
	fmt.Printf("Source: %s (Local File: %v)\n", manager.gtfsSource, manager.isLocalFile)
	fmt.Printf("Last Updated: %s\n", manager.lastUpdated)
	fmt.Println("Stops Count: ", len(manager.gtfsData.Stops))
	fmt.Println("Routes Count: ", len(manager.gtfsData.Routes))
	fmt.Println("Trips Count: ", len(manager.gtfsData.Trips))
	fmt.Println("Agencies Count: ", len(manager.gtfsData.Agencies))
}
