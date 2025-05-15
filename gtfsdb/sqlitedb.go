package gtfsdb

import (
	"database/sql"
	"fmt"
	"log"
)

// InitDB creates a new SQLite database with GTFS tables for agencies and routes
func InitDB(dbPath string) (*sql.DB, error) {
	// Open database connection
	db, err := sql.Open("sqlite", dbPath)
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

	// Create shapes table
	createTable(tx, "shapes", `
		CREATE TABLE IF NOT EXISTS shapes (
		    id INTEGER PRIMARY KEY AUTOINCREMENT,
		    shape_id TEXT NOT NULL,
		    lat REAL NOT NULL,
		    lon REAL NOT NULL,
		    shape_pt_sequence INTEGER NOT NULL
		);
    `)

	// Create stop_times table (junction table between trips and stops)
	createTable(tx, "stop_times", `
		CREATE TABLE IF NOT EXISTS stop_times (
			trip_id TEXT NOT NULL,
			arrival_time INTEGER NOT NULL,
			departure_time INTEGER NOT NULL,
			stop_id TEXT NOT NULL,
			stop_sequence INTEGER NOT NULL,
			stop_headsign TEXT,
			pickup_type INTEGER DEFAULT 0,
			drop_off_type INTEGER DEFAULT 0,
			shape_dist_traveled REAL,
			timepoint INTEGER DEFAULT 1,
			FOREIGN KEY (trip_id) REFERENCES trips(trip_id),
			FOREIGN KEY (stop_id) REFERENCES stops(stop_id),
			PRIMARY KEY (trip_id, stop_sequence)
		);
	`)
}

// createTable creates a table in the database
func createTable(tx *sql.Tx, tableName string, createStmt string) {
	_, err := tx.Exec(createStmt)
	if err != nil {
		log.Fatalf("Error creating table %s: %v", tableName, err)
	}
}
