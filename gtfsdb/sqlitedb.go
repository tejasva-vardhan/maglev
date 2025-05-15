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
	// Create agencies table
	createTable(tx, "agencies", `
		CREATE TABLE IF NOT EXISTS agencies (
			agency_id TEXT PRIMARY KEY,
			agency_name TEXT NOT NULL,
			agency_url TEXT NOT NULL,
			agency_timezone TEXT NOT NULL,
			agency_lang TEXT,
			agency_phone TEXT,
			agency_fare_url TEXT,
			agency_email TEXT
		);
	`)

	// Create routes table with foreign key to agencies
	createTable(tx, "routes", `
		CREATE TABLE IF NOT EXISTS routes (
			route_id TEXT PRIMARY KEY,
			agency_id TEXT,
			route_short_name TEXT,
			route_long_name TEXT,
			route_desc TEXT,
			route_type INTEGER NOT NULL,
			route_url TEXT,
			route_color TEXT,
			route_text_color TEXT,
			continuous_pickup INTEGER,
			continuous_drop_off INTEGER,
			FOREIGN KEY (agency_id) REFERENCES agencies(agency_id)
		);`)

	// First, create your regular stops table as you've already done
	createTable(tx, "stops", `
    CREATE TABLE IF NOT EXISTS stops (
        stop_id TEXT PRIMARY KEY,
        stop_code TEXT,
        stop_name TEXT NOT NULL,
        stop_desc TEXT,
        stop_lat REAL NOT NULL,
        stop_lon REAL NOT NULL,
        zone_id TEXT,
        stop_url TEXT,
        location_type INTEGER DEFAULT 0,
        stop_timezone TEXT,
        wheelchair_boarding INTEGER DEFAULT 0,
        level_id TEXT,
        platform_code TEXT
    );`,
	)

	// Then create an R*Tree virtual table that will serve as a spatial index
	createTable(tx, "stops_rtree", `
    CREATE VIRTUAL TABLE IF NOT EXISTS stops_rtree USING rtree(
        id,              -- Integer primary key for the R*Tree
        min_lat, max_lat, -- Latitude bounds
        min_lon, max_lon  -- Longitude bounds
    );`,
	)

	// You'll need to add a trigger to keep the R*Tree updated when stops are added
	createTable(tx, "stops_rtree_insert_trigger", `
    CREATE TRIGGER IF NOT EXISTS stops_rtree_insert_trigger
    AFTER INSERT ON stops
    BEGIN
        INSERT INTO stops_rtree(id, min_lat, max_lat, min_lon, max_lon)
        VALUES (new.rowid, new.stop_lat, new.stop_lat, new.stop_lon, new.stop_lon);
    END;`,
	)

	// Also add triggers for updates and deletes
	createTable(tx, "stops_rtree_update_trigger", `
    CREATE TRIGGER IF NOT EXISTS stops_rtree_update_trigger
    AFTER UPDATE ON stops
    BEGIN
        UPDATE stops_rtree SET
            min_lat = new.stop_lat,
            max_lat = new.stop_lat,
            min_lon = new.stop_lon,
            max_lon = new.stop_lon
        WHERE id = old.rowid;
    END;`,
	)

	createTable(tx, "stops_rtree_delete_trigger", `
    CREATE TRIGGER IF NOT EXISTS stops_rtree_delete_trigger
    AFTER DELETE ON stops
    BEGIN
        DELETE FROM stops_rtree WHERE id = old.rowid;
    END;`,
	)

	// Create calendar table (needed for trips references)
	createTable(tx, "calendar", `
		CREATE TABLE IF NOT EXISTS calendar (
			service_id TEXT PRIMARY KEY,
			monday INTEGER NOT NULL,
			tuesday INTEGER NOT NULL,
			wednesday INTEGER NOT NULL,
			thursday INTEGER NOT NULL,
			friday INTEGER NOT NULL,
			saturday INTEGER NOT NULL,
			sunday INTEGER NOT NULL,
			start_date TEXT NOT NULL,
			end_date TEXT NOT NULL
		);`,
	)

	// Create trips table
	createTable(tx, "trips", `
		CREATE TABLE IF NOT EXISTS trips (
			trip_id TEXT PRIMARY KEY,
			route_id TEXT NOT NULL,
			service_id TEXT NOT NULL,
			trip_headsign TEXT,
			trip_short_name TEXT,
			direction_id INTEGER,
			block_id TEXT,
			shape_id TEXT,
			wheelchair_accessible INTEGER DEFAULT 0,
			bikes_allowed INTEGER DEFAULT 0,
			FOREIGN KEY (route_id) REFERENCES routes(route_id),
			FOREIGN KEY (service_id) REFERENCES calendar(service_id)
		);
	`)

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
