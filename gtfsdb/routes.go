package gtfsdb

import (
	"database/sql"
	"fmt"
)

// Route represents a transit route in the GTFS feed
type Route struct {
	ID                string // route_id
	AgencyID          string // agency_id
	ShortName         string // route_short_name
	LongName          string // route_long_name
	Desc              string // route_desc
	Type              int    // route_type
	URL               string // route_url
	Color             string // route_color
	TextColor         string // route_text_color
	ContinuousPickup  int    // continuous_pickup
	ContinuousDropOff int    // continuous_drop_off
}

// InsertRoute adds a new route to the database
func InsertRoute(db *sql.DB, route Route) error {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO routes (
			route_id, agency_id, route_short_name, route_long_name,
			route_desc, route_type, route_url, route_color,
			route_text_color, continuous_pickup, continuous_drop_off
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`,
		route.ID, route.AgencyID, route.ShortName, route.LongName,
		route.Desc, route.Type, route.URL, route.Color,
		route.TextColor, route.ContinuousPickup, route.ContinuousDropOff,
	)
	if err != nil {
		return fmt.Errorf("error inserting route: %w", err)
	}
	return nil
}

// createRoutesTable creates the "routes" table in the database using the provided transaction object.
// It defines fields for storing route information and enforces a foreign key constraint on the "agency_id" column.
func createRoutesTable(tx *sql.Tx) {
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
}
