package gtfsdb

// Shape represents points that define a vehicle's path
type Shape struct {
	ID       string  // shape_id
	Lat      float64 // shape_pt_lat
	Lon      float64 // shape_pt_lon
	Sequence int     // shape_pt_sequence
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
