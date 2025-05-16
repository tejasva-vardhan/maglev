package gtfsdb

import (
	"database/sql"
	"fmt"
)

// Stop represents a transit stop or station in the GTFS feed
type Stop struct {
	ID                 string  // stop_id
	Code               string  // stop_code
	Name               string  // stop_name
	Desc               string  // stop_desc
	Lat                float64 // stop_lat
	Lon                float64 // stop_lon
	ZoneID             string  // zone_id
	URL                string  // stop_url
	LocationType       int     // location_type
	Timezone           string  // stop_timezone
	WheelchairBoarding int     // wheelchair_boarding
	LevelID            string  // level_id
	PlatformCode       string  // platform_code
}

// InsertStops add new stops to the database
func InsertStops(db *sql.DB, stops []Stop) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO stops (
			stop_id, stop_code, stop_name, stop_desc, stop_lat, stop_lon,
			zone_id, stop_url, location_type, stop_timezone,
			wheelchair_boarding, level_id, platform_code
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`)
	if err != nil {
		tx.Rollback() // nolint:errcheck
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close() // nolint:errcheck

	for _, stop := range stops {
		_, err := stmt.Exec(
			stop.ID, stop.Code, stop.Name, stop.Desc, stop.Lat, stop.Lon,
			stop.ZoneID, stop.URL, stop.LocationType, stop.Timezone,
			stop.WheelchairBoarding, stop.LevelID, stop.PlatformCode,
		)
		if err != nil {
			tx.Rollback() // nolint:errcheck
			return fmt.Errorf("error inserting stop: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func createStopsTable(tx *sql.Tx) {
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
}
