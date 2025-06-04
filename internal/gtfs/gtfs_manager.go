package gtfs

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"maglev.onebusaway.org/gtfsdb"
	"maglev.onebusaway.org/internal/utils"

	"github.com/jamespfennell/gtfs"
	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// Manager manages the GTFS data and provides methods to access it
type Manager struct {
	gtfsSource       string
	gtfsData         *gtfs.Static
	GtfsDB           *gtfsdb.Client
	lastUpdated      time.Time
	isLocalFile      bool
	realTimeTrips    []gtfs.Trip
	realTimeVehicles []gtfs.Vehicle
	realTimeMutex    sync.RWMutex
	config           Config
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
		isLocalFile: isLocalFile,
		config:      config,
	}
	manager.setStaticGTFS(staticData)

	gtfsDB, err := buildGtfsDB(config, isLocalFile)
	if err != nil {
		return nil, fmt.Errorf("error building GTFS database: %w", err)
	}
	manager.GtfsDB = gtfsDB

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

func (manager *Manager) GetAgencies() []gtfs.Agency {
	return manager.gtfsData.Agencies
}

func (manager *Manager) GetTrips() []gtfs.ScheduledTrip {
	return manager.gtfsData.Trips
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

// RoutesForAgencyID retrieves all routes associated with the specified agency ID from the GTFS data.
func (manager *Manager) RoutesForAgencyID(agencyID string) []*gtfs.Route {
	var agencyRoutes []*gtfs.Route

	for i := range manager.gtfsData.Routes {
		if manager.gtfsData.Routes[i].Agency.Id == agencyID {
			agencyRoutes = append(agencyRoutes, &manager.gtfsData.Routes[i])
		}
	}

	return agencyRoutes
}

type stopWithDistance struct {
	stop     *gtfs.Stop
	distance float64
}

func (manager *Manager) GetStopsForLocation(lat, lon float64, radius float64, latSpan, lonSpan float64, query string, maxCount int) []*gtfs.Stop {
	const epsilon = 1e-6

	if radius == 0 {
		radius = 1000
	}
	if query != "" {
		radius *= 10
	}

	var candidates []stopWithDistance

	for i := range manager.gtfsData.Stops {
		stop := &manager.gtfsData.Stops[i]
		if stop.Latitude == nil || stop.Longitude == nil {
			continue
		}

		if query != "" {
			if stop.Code == query {
				distance := utils.Haversine(lat, lon, *stop.Latitude, *stop.Longitude)
				if distance <= radius {
					return []*gtfs.Stop{stop}
				}
			}
			continue
		}

		if latSpan > 0 && lonSpan > 0 {
			if math.Abs(*stop.Latitude-lat) <= latSpan+epsilon && math.Abs(*stop.Longitude-lon) <= lonSpan+epsilon {
				distance := utils.Haversine(lat, lon, *stop.Latitude, *stop.Longitude)
				candidates = append(candidates, stopWithDistance{stop, distance})
			}
		} else {
			distance := utils.Haversine(lat, lon, *stop.Latitude, *stop.Longitude)
			if distance <= radius {
				candidates = append(candidates, stopWithDistance{stop, distance})
			}
		}
	}

	// Sort by distance
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].distance < candidates[j].distance
	})

	// Limit to maxCount
	var stops []*gtfs.Stop
	for i := 0; i < len(candidates) && i < maxCount; i++ {
		stops = append(stops, candidates[i].stop)
	}

	return stops
}

func (manager *Manager) VehiclesForAgencyID(agencyID string) []gtfs.Vehicle {
	routes := manager.RoutesForAgencyID(agencyID)
	routeIDs := make(map[string]bool) // all route IDs for the agency.
	for _, route := range routes {
		routeIDs[route.Id] = true
	}

	var vehicles []gtfs.Vehicle
	for _, v := range manager.GetRealTimeVehicles() {
		if v.Trip != nil {
			if routeIDs[v.Trip.ID.RouteID] {
				vehicles = append(vehicles, v)
			}
		}
	}

	return vehicles
}

func (manager *Manager) PrintStatistics() {
	fmt.Printf("Source: %s (Local File: %v)\n", manager.gtfsSource, manager.isLocalFile)
	fmt.Printf("Last Updated: %s\n", manager.lastUpdated)
	fmt.Println("Stops Count: ", len(manager.gtfsData.Stops))
	fmt.Println("Routes Count: ", len(manager.gtfsData.Routes))
	fmt.Println("Trips Count: ", len(manager.gtfsData.Trips))
	fmt.Println("Agencies Count: ", len(manager.gtfsData.Agencies))
}
