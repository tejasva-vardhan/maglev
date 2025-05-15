package gtfsdb

import (
	"database/sql"
	"fmt"
)

// InsertStopTimeBatch inserts multiple stop times using a transaction for better performance
func InsertStopTimeBatch(db *sql.DB, stopTimes []StopTime) error {
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
