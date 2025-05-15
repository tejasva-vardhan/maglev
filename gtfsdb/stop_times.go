package gtfsdb

import (
	"database/sql"
	"fmt"
)

// StopTime represents a vehicle arrival/departure at a specific stop in the GTFS feed
type StopTime struct {
	TripID        string // trip_id
	ArrivalTime   int    // arrival_time (HH:MM:SS)
	DepartureTime int    // departure_time (HH:MM:SS)
	StopID        string // stop_id
	StopSequence  int    // stop_sequence
	StopHeadsign  string // stop_headsign
	PickupType    int    // pickup_type
	DropOffType   int    // drop_off_type
	Timepoint     int    // timepoint
}

// InsertStopTimes inserts multiple stop times using a transaction for better performance
func InsertStopTimes(db *sql.DB, stopTimes []StopTime) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO stop_times (
			trip_id, arrival_time, departure_time, stop_id, stop_sequence,
			stop_headsign, pickup_type, drop_off_type, timepoint
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
	`)
	if err != nil {
		tx.Rollback() // nolint:errcheck
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close() // nolint:errcheck

	for _, st := range stopTimes {
		_, err := stmt.Exec(
			st.TripID, st.ArrivalTime, st.DepartureTime, st.StopID, st.StopSequence,
			st.StopHeadsign, st.PickupType, st.DropOffType, st.Timepoint,
		)
		if err != nil {
			tx.Rollback() // nolint:errcheck
			return fmt.Errorf("error inserting stop_time: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// createStopTimesTable creates the "stop_times" table in the database using the provided transaction object.
// The table stores information about stops within trips, including fields for arrival time, departure time, and stop sequence.
// It includes foreign key constraints referencing "trips" and "stops" tables and a primary key on trip_id and stop_sequence.
func createStopTimesTable(tx *sql.Tx) {
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
