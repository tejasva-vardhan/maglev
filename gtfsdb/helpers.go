package gtfsdb

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"github.com/jamespfennell/gtfs"
	"log"
	"maglev.onebusaway.org/internal/appconf"
	"strings"
	"time"
)

//go:embed schema.sql
var ddl string

// createDB creates a new SQLite database with tables for static GTFS data
func createDB(config Config) (*sql.DB, error) {
	if config.Env == appconf.Test && config.DBPath != ":memory:" {
		log.Fatal("DB is being created in a file.", config.DBPath)
	}

	db, err := sql.Open("sqlite", config.DBPath)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	err = performDatabaseMigration(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("error performing database migration: %w", err)
	}

	return db, nil
}

func performDatabaseMigration(ctx context.Context, db *sql.DB) error {
	statements := strings.Split(ddl, "-- migrate") // Split DDL into individual statements
	for _, stmt := range statements {
		trimmedStmt := strings.TrimSpace(stmt)
		if trimmedStmt == "" {
			continue // Skip empty statements
		}
		if _, err := db.ExecContext(ctx, trimmedStmt); err != nil {
			return fmt.Errorf("error executing DDL statement [%s]: %w", trimmedStmt, err)
		}
	}
	return nil
}

func (c *Client) processAndStoreGTFSData(b []byte) error {
	startTime := time.Now()
	defer func() {
		endTime := time.Now()

		c.importRuntime = endTime.Sub(startTime)

		if c.config.verbose {
			log.Println("Importing GTFS data took", c.importRuntime.String())
		}
	}()
	
	var staticCounts map[string]int

	staticData, err := gtfs.ParseStatic(b, gtfs.ParseStaticOptions{})
	if err != nil {
		return err
	}

	if c.config.verbose {
		fmt.Printf("retrieved static data (warnings: %d)\n", len(staticData.Warnings))
		fmt.Print("========\n\n")

		staticCounts = c.staticDataCounts(staticData)
		for k, v := range staticCounts {
			fmt.Printf("%s: %d\n", k, v)
		}

		fmt.Print("========\n\n")
	}

	ctx := context.Background()

	for _, a := range staticData.Agencies {
		params := CreateAgencyParams{
			ID:       a.Id,
			Name:     a.Name,
			Url:      a.Url,
			Timezone: a.Timezone,
			Lang:     toNullString(a.Language),
			Phone:    toNullString(a.Phone),
			FareUrl:  toNullString(a.FareUrl),
			Email:    toNullString(a.Email),
		}

		_, err := c.Queries.CreateAgency(ctx, params)
		if err != nil {
			log.Fatal("Unable to create agency", err)
		}
	}

	singleAgencyID := ""
	if len(staticData.Agencies) == 1 {
		singleAgencyID = staticData.Agencies[0].Id
	}

	for _, r := range staticData.Routes {
		route := CreateRouteParams{
			ID:                r.Id,
			AgencyID:          pickFirstAvailable(r.Agency.Id, singleAgencyID),
			ShortName:         toNullString(r.ShortName),
			LongName:          toNullString(r.LongName),
			Desc:              toNullString(r.Description),
			Type:              int64(r.Type),
			Url:               toNullString(r.Url),
			Color:             toNullString(r.Color),
			TextColor:         toNullString(r.TextColor),
			ContinuousPickup:  toNullInt64(int64(r.ContinuousPickup)),
			ContinuousDropOff: toNullInt64(int64(r.ContinuousDropOff)),
		}

		_, err := c.Queries.CreateRoute(ctx, route)

		if err != nil {
			log.Fatal("Unable to create route: ", err)
		}
	}

	var allStopParams []CreateStopParams
	for _, s := range staticData.Stops {
		params := CreateStopParams{
			ID:                 s.Id,
			Code:               toNullString(s.Code),
			Name:               toNullString(s.Name),
			Desc:               toNullString(s.Description),
			Lat:                *s.Latitude,
			Lon:                *s.Longitude,
			ZoneID:             toNullString(s.ZoneId),
			Url:                toNullString(s.Url),
			LocationType:       toNullInt64(int64(s.Type)),
			Timezone:           toNullString(s.Timezone),
			WheelchairBoarding: toNullInt64(int64(s.WheelchairBoarding)),
			PlatformCode:       toNullString(s.PlatformCode),
		}

		allStopParams = append(allStopParams, params)
	}
	err = c.bulkInsertStops(ctx, allStopParams)
	if err != nil {
		log.Fatalf("Unable to create stops: %v\n", err)
	}

	for _, s := range staticData.Services {
		params := CreateCalendarParams{
			ID:        s.Id,
			Monday:    boolToInt(s.Monday),
			Tuesday:   boolToInt(s.Tuesday),
			Wednesday: boolToInt(s.Wednesday),
			Thursday:  boolToInt(s.Thursday),
			Friday:    boolToInt(s.Friday),
			Saturday:  boolToInt(s.Saturday),
			Sunday:    boolToInt(s.Sunday),
			StartDate: s.StartDate.Format("20060102"),
			EndDate:   s.EndDate.Format("20060102"),
		}

		_, err := c.Queries.CreateCalendar(ctx, params)
		if err != nil {
			log.Fatal("Unable to create calendar: ", err)
		}
	}

	var allTripParams []CreateTripParams
	for _, t := range staticData.Trips {
		params := CreateTripParams{
			ID:                   t.ID,
			RouteID:              t.Route.Id,
			ServiceID:            t.Service.Id,
			TripHeadsign:         toNullString(t.Headsign),
			TripShortName:        toNullString(t.ShortName),
			DirectionID:          toNullInt64(int64(t.DirectionId)),
			BlockID:              toNullString(t.BlockID),
			ShapeID:              toNullString(t.Shape.ID),
			WheelchairAccessible: toNullInt64(int64(t.WheelchairAccessible)),
			BikesAllowed:         toNullInt64(int64(t.BikesAllowed)),
		}
		allTripParams = append(allTripParams, params)
	}
	err = c.bulkInsertTrips(ctx, allTripParams)
	if err != nil {
		log.Fatalf("Unable to create trips: %v\n", err)
	}

	var allStopTimeParams []CreateStopTimeParams
	for _, t := range staticData.Trips {
		for _, st := range t.StopTimes {
			params := CreateStopTimeParams{
				TripID:        t.ID,
				ArrivalTime:   int64(st.ArrivalTime),
				DepartureTime: int64(st.DepartureTime),
				StopID:        st.Stop.Id,
				StopSequence:  int64(st.StopSequence),
				StopHeadsign:  toNullString(st.Headsign),
				PickupType:    toNullInt64(int64(st.PickupType)),
				DropOffType:   toNullInt64(int64(st.DropOffType)),
				Timepoint:     toNullInt64(boolToInt(st.ExactTimes)),
			}

			allStopTimeParams = append(allStopTimeParams, params)
		}
	}
	err = c.bulkInsertStopTimes(ctx, allStopTimeParams)
	if err != nil {
		log.Fatalf("Unable to create stop times: %v\n", err)
	}

	var allShapeParams []CreateShapeParams
	for _, s := range staticData.Shapes {
		for idx, pt := range s.Points {
			params := CreateShapeParams{
				ShapeID:         s.ID,
				Lat:             pt.Latitude,
				Lon:             pt.Longitude,
				ShapePtSequence: int64(idx),
			}
			allShapeParams = append(allShapeParams, params)
		}
	}
	err = c.bulkInsertShapes(ctx, allShapeParams)
	if err != nil {
		log.Fatalf("Unable to create shapes: %v\n", err)
	}

	if c.config.verbose {
		counts, err := c.TableCounts()
		if err != nil {
			log.Fatalf("Failed to get table counts: %v", err)
		}
		for k, v := range counts {
			fmt.Printf("%s: %d (Static matches? %v)\n", k, v, v == staticCounts[k])
		}
	}

	return nil
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func toNullInt64(i int64) sql.NullInt64 {
	if i != 0 {
		return sql.NullInt64{
			Int64: i,
			Valid: true,
		}
	}
	return sql.NullInt64{}
}

// toNullString converts a string to sql.NullString
func toNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}

func pickFirstAvailable(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func (c *Client) bulkInsertStops(ctx context.Context, stops []CreateStopParams) error {
	db := c.DB
	queries := c.Queries

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // nolint:errcheck

	qtx := queries.WithTx(tx)
	for _, params := range stops {
		_, err := qtx.CreateStop(ctx, params)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (c *Client) bulkInsertTrips(ctx context.Context, trips []CreateTripParams) error {
	db := c.DB
	queries := c.Queries

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // nolint:errcheck

	qtx := queries.WithTx(tx)
	for _, params := range trips {
		_, err := qtx.CreateTrip(ctx, params)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (c *Client) bulkInsertStopTimes(ctx context.Context, stopTimes []CreateStopTimeParams) error {
	db := c.DB
	queries := c.Queries

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // nolint:errcheck

	qtx := queries.WithTx(tx)
	for _, params := range stopTimes {
		_, err := qtx.CreateStopTime(ctx, params)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (c *Client) bulkInsertShapes(ctx context.Context, shapes []CreateShapeParams) error {
	db := c.DB
	queries := c.Queries

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // nolint:errcheck

	qtx := queries.WithTx(tx)
	for _, params := range shapes {
		_, err := qtx.CreateShape(ctx, params)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
