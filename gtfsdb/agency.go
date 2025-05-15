package gtfsdb

import (
	"context"
	"database/sql"
	"fmt"
)

// Agency represents a transit agency in the GTFS feed
type Agency struct {
	Id       string // agency_id
	Name     string // agency_name
	Url      string // agency_url
	Timezone string // agency_timezone
	Language string // agency_lang
	Phone    string // agency_phone
	FareUrl  string // agency_fare_url
	Email    string // agency_email
}

// QueryAgencies retrieves a list of transit agencies from the database and returns them as a slice of Agency objects.
func (c *Client) QueryAgencies(ctx context.Context) ([]Agency, error) {
	rows, err := c.DB.QueryContext(
		ctx,
		`SELECT agency_id, agency_name, agency_url, agency_timezone,
				agency_lang, agency_phone, agency_fare_url, agency_email
				FROM agencies`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // nolint:errcheck

	var agencies []Agency
	for rows.Next() {
		var agency Agency
		err := rows.Scan(&agency.Id, &agency.Name, &agency.Url, &agency.Timezone,
			&agency.Language, &agency.Phone, &agency.FareUrl, &agency.Email,
		)
		if err != nil {
			return nil, err
		}
		agencies = append(agencies, agency)
	}

	return agencies, nil
}

// insertAgency adds a new agency to the database
func insertAgency(db *sql.DB, agency Agency) error {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO agencies (
			agency_id, agency_name, agency_url, agency_timezone,
			agency_lang, agency_phone, agency_fare_url, agency_email
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?);
	`,
		agency.Id, agency.Name, agency.Url, agency.Timezone,
		agency.Language, agency.Phone, agency.FareUrl, agency.Email,
	)
	if err != nil {
		return fmt.Errorf("error inserting agency: %w", err)
	}
	return nil
}
