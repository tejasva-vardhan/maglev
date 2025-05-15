package gtfsdb

import (
	"database/sql"
	"fmt"
)

// InsertStopBatch add new stops to the database
func InsertStopBatch(db *sql.DB, stops []Stop) error {
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

func InsertTripBatch(db *sql.DB, trips []Trip) error {
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

// InsertShapes inserts multiple stop times using a transaction for better performance
func InsertShapes(db *sql.DB, shapes []Shape) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO shapes (
			shape_id, lat, lon, shape_pt_sequence
		) VALUES (?, ?, ?, ?);
	`)
	if err != nil {
		tx.Rollback() // nolint:errcheck
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close() // nolint:errcheck

	for _, st := range shapes {
		_, err := stmt.Exec(
			st.ID, st.Lat, st.Lon, st.Sequence,
		)
		if err != nil {
			tx.Rollback() // nolint:errcheck
			return fmt.Errorf("error inserting shapes: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

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
