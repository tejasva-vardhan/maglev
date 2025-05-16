package gtfsdb

import (
	"database/sql"
	"fmt"
)

// Calendar represents service dates for trips in the GTFS feed
type Calendar struct {
	ServiceID string // service_id
	Monday    int    // monday
	Tuesday   int    // tuesday
	Wednesday int    // wednesday
	Thursday  int    // thursday
	Friday    int    // friday
	Saturday  int    // saturday
	Sunday    int    // sunday
	StartDate string // start_date (YYYYMMDD)
	EndDate   string // end_date (YYYYMMDD)
}

// InsertCalendar adds a new calendar entry to the database
func InsertCalendar(db *sql.DB, calendar Calendar) error {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO calendar (
			service_id, monday, tuesday, wednesday, thursday,
			friday, saturday, sunday, start_date, end_date
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`,
		calendar.ServiceID, calendar.Monday, calendar.Tuesday, calendar.Wednesday, calendar.Thursday,
		calendar.Friday, calendar.Saturday, calendar.Sunday, calendar.StartDate, calendar.EndDate,
	)
	if err != nil {
		return fmt.Errorf("error inserting calendar: %w", err)
	}
	return nil
}

func createCalendarTable(tx *sql.Tx) {
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
}
