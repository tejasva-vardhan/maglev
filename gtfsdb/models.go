package gtfsdb

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

// Shape represents points that define a vehicle's path
type Shape struct {
	ID       string  // shape_id
	Lat      float64 // shape_pt_lat
	Lon      float64 // shape_pt_lon
	Sequence int     // shape_pt_sequence
}

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
