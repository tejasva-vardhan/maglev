package gtfs

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jamespfennell/gtfs"
)

// Manager manages the GTFS data and provides methods to access it
type Manager struct {
	gtfsSource  string
	gtfsData    *gtfs.Static
	lastUpdated time.Time
	isLocalFile bool
}

// InitGTFSManager initializes the Manager with the GTFS data from the given source
// The source can be either a URL or a local file path
func InitGTFSManager(gtfsSource string) (*Manager, error) {
	isLocalFile := !strings.HasPrefix(gtfsSource, "http://") && !strings.HasPrefix(gtfsSource, "https://")

	staticData, err := loadGTFSData(gtfsSource, isLocalFile)
	if err != nil {
		return nil, err
	}

	return &Manager{
		gtfsSource:  gtfsSource,
		gtfsData:    staticData,
		lastUpdated: time.Now(),
		isLocalFile: isLocalFile,
	}, nil
}

// loadGTFSData loads and parses GTFS data from either a URL or a local file
func loadGTFSData(source string, isLocalFile bool) (*gtfs.Static, error) {
	var b []byte
	var err error

	if isLocalFile {
		b, err = os.ReadFile(source)
		if err != nil {
			return nil, fmt.Errorf("error reading local GTFS file: %w", err)
		}
	} else {
		resp, err := http.Get(source)
		if err != nil {
			return nil, fmt.Errorf("error downloading GTFS data: %w", err)
		}
		defer resp.Body.Close() // nolint

		b, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading GTFS data: %w", err)
		}
	}

	staticData, err := gtfs.ParseStatic(b, gtfs.ParseStaticOptions{})
	if err != nil {
		return nil, fmt.Errorf("error parsing GTFS data: %w", err)
	}

	return staticData, nil
}

// UpdateGTFSPeriodically updates the GTFS data on a regular schedule
// Only updates if the source is a URL, not a local file
func (manager *Manager) updateGTFSPeriodically() { // nolint
	// If it's a local file, don't update periodically
	if manager.isLocalFile {
		log.Printf("GTFS source is a local file, skipping periodic updates")
		return
	}

	// Update every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for { // nolint
		select {
		case <-ticker.C:
			// Create a context with timeout for the download
			_, cancel := context.WithTimeout(context.Background(), 60*time.Second)

			// Download and parse the GTFS feed
			staticData, err := loadGTFSData(manager.gtfsSource, false)
			cancel() // Always cancel the context when done

			if err != nil {
				// Log error but don't crash the application
				log.Printf("Error updating GTFS data: %v", err)
				continue
			}

			// Update the GTFS data in the manager
			manager.gtfsData = staticData
			manager.lastUpdated = time.Now()

			log.Printf("GTFS data updated successfully for %v", manager.gtfsSource)
		}
	}
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

func (manager *Manager) FindAgency(id string) *gtfs.Agency {
	for _, agency := range manager.gtfsData.Agencies {
		if agency.Id == id {
			return &agency
		}
	}
	return nil
}

func (manager *Manager) PrintStatistics() {
	fmt.Printf("Source: %s (Local File: %v)\n", manager.gtfsSource, manager.isLocalFile)
	fmt.Printf("Last Updated: %s\n", manager.lastUpdated)
	fmt.Println("Stops Count: ", len(manager.gtfsData.Stops))
	fmt.Println("Routes Count: ", len(manager.gtfsData.Routes))
	fmt.Println("Trips Count: ", len(manager.gtfsData.Trips))
	fmt.Println("Agencies Count: ", len(manager.gtfsData.Agencies))
}
