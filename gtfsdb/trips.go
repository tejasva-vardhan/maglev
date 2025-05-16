package gtfsdb

import (
	"database/sql"
	"fmt"
)

// Trip represents a journey made by a vehicle in the GTFS feed
type Trip struct {
	ID                   string // trip_id
	RouteID              string // route_id
	ServiceID            string // service_id
	Headsign             string // trip_headsign
	ShortName            string // trip_short_name
	DirectionID          int    // direction_id
	BlockID              string // block_id
	ShapeID              string // shape_id
	WheelchairAccessible int    // wheelchair_accessible
	BikesAllowed         int    // bikes_allowed
}

func InsertTrips(db *sql.DB, trips []Trip) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO trips (
			trip_id, route_id, service_id, trip_headsign, trip_short_name,
			direction_id, block_id, shape_id, wheelchair_accessible, bikes_allowed
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`)
	if err != nil {
		tx.Rollback() // nolint:errcheck
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close() // nolint:errcheck

	for _, trip := range trips {
		_, err := stmt.Exec(
			trip.ID, trip.RouteID, trip.ServiceID, trip.Headsign, trip.ShortName,
			trip.DirectionID, trip.BlockID, trip.ShapeID, trip.WheelchairAccessible, trip.BikesAllowed,
		)
		if err != nil {
			tx.Rollback() // nolint:errcheck
			return fmt.Errorf("error inserting trip: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func createTripsTable(tx *sql.Tx) {
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
}
