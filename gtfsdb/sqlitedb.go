package gtfsdb

import (
	"database/sql"
	"fmt"
	"log"
	"maglev.onebusaway.org/internal/appconf"
)

// InitDB creates a new SQLite database with GTFS tables for agencies and routes
func InitDB(config Config) (*sql.DB, error) {
	if config.Env == appconf.Test && config.DBPath != ":memory:" {
		log.Fatal("DB is being created in a file.", config.DBPath)
	}

	// Open database connection
	db, err := sql.Open("sqlite", config.DBPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, fmt.Errorf("error enabling foreign keys: %w", err)
	}

	// Create tables within a transaction
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}

	createTables(tx)

	// Create indexes for better performance
	_, err = tx.Exec(`
		CREATE INDEX IF NOT EXISTS idx_routes_agency_id ON routes(agency_id);
		CREATE INDEX IF NOT EXISTS idx_trips_route_id ON trips(route_id);
		CREATE INDEX IF NOT EXISTS idx_trips_service_id ON trips(service_id);
		CREATE INDEX IF NOT EXISTS idx_stop_times_trip_id ON stop_times(trip_id);
		CREATE INDEX IF NOT EXISTS idx_stop_times_stop_id ON stop_times(stop_id);
	`)
	if err != nil {
		tx.Rollback() // nolint:errcheck
		log.Fatalf("error creating indexes: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return db, nil
}

func createTables(tx *sql.Tx) {
	createAgenciesTable(tx)
	createRoutesTable(tx)
	createStopsTable(tx)
	createCalendarTable(tx)
	createTripsTable(tx)
	createShapesTable(tx)
	createStopTimesTable(tx)
}

// createTable creates a table in the database
func createTable(tx *sql.Tx, tableName string, createStmt string) {
	_, err := tx.Exec(createStmt)
	if err != nil {
		log.Fatalf("Error creating table %s: %v", tableName, err)
	}
}
