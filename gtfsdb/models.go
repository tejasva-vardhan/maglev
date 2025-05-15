package gtfsdb

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
