package gtfs

import (
	"context"
	"fmt"
	"github.com/jamespfennell/gtfs"
	"io"
	"log"
	"maglev.onebusaway.org/gtfsdb"
	"net/http"
	"os"
	"time"
)

func rawGtfsData(source string, isLocalFile bool) ([]byte, error) {
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
	return b, nil
}

func buildGtfsDB(config Config, isLocalFile bool) (*gtfsdb.Client, error) {
	dbConfig := gtfsdb.NewConfig(config.GTFSDataPath, config.Env, config.Verbose)
	client, err := gtfsdb.NewClient(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create GTFS database client: %w", err)
	}

	ctx := context.Background()

	if isLocalFile {
		err = client.ImportFromFile(ctx, config.GtfsURL)
	} else {
		err = client.DownloadAndStore(ctx, config.GtfsURL)
	}

	return client, err
}

// loadGTFSData loads and parses GTFS data from either a URL or a local file
func loadGTFSData(source string, isLocalFile bool) (*gtfs.Static, error) {
	b, err := rawGtfsData(source, isLocalFile)
	if err != nil {
		return nil, fmt.Errorf("error reading GTFS data: %w", err)
	}

	staticData, err := gtfs.ParseStatic(b, gtfs.ParseStaticOptions{})
	if err != nil {
		return nil, fmt.Errorf("error parsing GTFS data: %w", err)
	}

	return staticData, nil
}

// UpdateGTFSPeriodically updates the GTFS data on a regular schedule
// Only updates if the source is a URL, not a local file
func (manager *Manager) updateStaticGTFS() { // nolint
	defer manager.wg.Done()

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
			manager.setStaticGTFS(staticData)
		case <-manager.shutdownChan:
			log.Println("Shutting down static GTFS updates")
			return
		}
	}
}

func (manager *Manager) setStaticGTFS(staticData *gtfs.Static) {
	manager.gtfsData = staticData
	manager.lastUpdated = time.Now()

	// perform post-processing here!

	if manager.config.Verbose {
		log.Printf("GTFS data updated successfully for %v", manager.gtfsSource)
	}
}
