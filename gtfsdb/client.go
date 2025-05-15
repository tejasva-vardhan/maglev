package gtfsdb

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jamespfennell/gtfs"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Client is the main entry point for the library
type Client struct {
	config        Config
	DB            *sql.DB
	importRuntime time.Duration
}

// NewClient creates a new Client with the provided configuration
func NewClient(config Config) *Client {
	db, err := InitDB(config.DBPath)
	if err != nil {
		log.Fatal("Unable to create DB", err)
	} else if config.verbose {
		log.Println("Successfully created tables")
	}

	client := &Client{
		config: config,
		DB:     db,
	}
	return client
}

func (c *Client) Close() error {
	return c.DB.Close()
}

// DownloadAndStore downloads GTFS data from the given URL and stores it in the database
func (c *Client) DownloadAndStore(ctx context.Context, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = c.processAndStoreGTFSData(b)

	return err
}

func (c *Client) processAndStoreGTFSData(b []byte) error {
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

	for _, a := range staticData.Agencies {
		agency := Agency{
			Id:       a.Id,
			Name:     a.Name,
			Url:      a.Url,
			Timezone: a.Timezone,
			Language: a.Language,
			Phone:    a.Phone,
			FareUrl:  a.FareUrl,
			Email:    a.Email,
		}
		err := insertAgency(c.DB, agency)

		if err != nil {
			log.Fatal("Unable to create agency", err)
		}
	}

	for _, r := range staticData.Routes {
		route := Route{
			ID:                r.Id,
			AgencyID:          r.Agency.Id,
			ShortName:         r.ShortName,
			LongName:          r.LongName,
			Desc:              r.Description,
			Type:              int(r.Type),
			URL:               r.Url,
			Color:             r.Color,
			TextColor:         r.TextColor,
			ContinuousPickup:  int(r.ContinuousPickup),
			ContinuousDropOff: int(r.ContinuousDropOff),
		}

		err := InsertRoute(c.DB, route)

		if err != nil {
			log.Fatal("Unable to create route", err)
		}
	}

	var allStops []Stop
	for _, s := range staticData.Stops {
		stop := Stop{
			ID:                 s.Id,
			Code:               s.Code,
			Name:               s.Name,
			Desc:               s.Description,
			Lat:                *s.Latitude,
			Lon:                *s.Longitude,
			ZoneID:             s.ZoneId,
			URL:                s.Url,
			LocationType:       int(s.Type),
			Timezone:           s.Timezone,
			WheelchairBoarding: 0,
			LevelID:            "",
			PlatformCode:       "",
		}
		allStops = append(allStops, stop)
	}
	err = InsertStopBatch(c.DB, allStops)
	if err != nil {
		log.Fatal("Unable to create stops\n", err)
	}

	for _, s := range staticData.Services {
		cal := Calendar{
			ServiceID: s.Id,
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

		err := InsertCalendar(c.DB, cal)
		if err != nil {
			log.Fatal("Unable to create calendar\n", err, cal)
		}
	}

	var allTrips []Trip
	for _, t := range staticData.Trips {
		trip := Trip{
			ID:                   t.ID,
			RouteID:              t.Route.Id,
			ServiceID:            t.Service.Id,
			Headsign:             t.Headsign,
			ShortName:            t.ShortName,
			DirectionID:          int(t.DirectionId),
			BlockID:              t.BlockID,
			ShapeID:              "",
			WheelchairAccessible: int(t.WheelchairAccessible),
			BikesAllowed:         int(t.BikesAllowed),
		}
		allTrips = append(allTrips, trip)
	}
	err = InsertTripBatch(c.DB, allTrips)
	if err != nil {
		log.Fatal("Unable to create trips\n", err)
	}

	var allStopTimes []StopTime
	for _, t := range staticData.Trips {
		for _, st := range t.StopTimes {
			stopTime := StopTime{
				TripID:        t.ID,
				ArrivalTime:   int(st.ArrivalTime),
				DepartureTime: int(st.DepartureTime),
				StopID:        st.Stop.Id,
				StopSequence:  st.StopSequence,
				PickupType:    int(st.PickupType),
				DropOffType:   int(st.DropOffType),
				Timepoint:     boolToInt(st.ExactTimes),
			}
			allStopTimes = append(allStopTimes, stopTime)
		}
	}
	err = InsertStopTimeBatch(c.DB, allStopTimes)
	if err != nil {
		log.Fatal("Unable to create stop times\n", err)
	}

	var allShapes []Shape
	for _, s := range staticData.Shapes {
		for idx, pt := range s.Points {
			shape := Shape{
				ID:       s.ID,
				Lat:      pt.Latitude,
				Lon:      pt.Longitude,
				Sequence: idx,
			}
			allShapes = append(allShapes, shape)
		}
	}
	err = InsertShapes(c.DB, allShapes)
	if err != nil {
		log.Fatal("Unable to create stop times\n", err)
	}

	counts, err := c.TableCounts()
	if err != nil {
		log.Fatalf("Failed to get table counts: %v", err)
	}

	if c.config.verbose {
		for k, v := range counts {
			fmt.Printf("%s: %d (Static matches? %v)\n", k, v, v == staticCounts[k])
		}
	}

	return nil
}

// ImportFromFile imports GTFS data from a local zip file into the database
func (c *Client) ImportFromFile(ctx context.Context, path string) error {
	startTime := time.Now()
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = c.processAndStoreGTFSData(data)
	endTime := time.Now()

	c.importRuntime = endTime.Sub(startTime)

	if c.config.verbose {
		log.Println("Importing GTFS data took", c.importRuntime.String())
	}

	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
