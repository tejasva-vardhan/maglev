package models

type VehicleStatus struct {
	VehicleID              string      `json:"vehicleId"`
	LastLocationUpdateTime int64       `json:"lastLocationUpdateTime,omitempty"`
	LastUpdateTime         int64       `json:"lastUpdateTime,omitempty"`
	Location               *Location   `json:"location,omitempty"`
	Status                 string      `json:"status,omitempty"`
	Phase                  string      `json:"phase,omitempty"`
	TripStatus             *TripStatus `json:"tripStatus,omitempty"`
}

type Location struct {
	Lat float32 `json:"lat"`
	Lon float32 `json:"lon"`
}

type TripStatus struct {
	ServiceDate                int64    `json:"serviceDate"`
	ActiveTripID               string   `json:"activeTripId"`
	Phase                      string   `json:"phase"`
	Status                     string   `json:"status"`
	Predicted                  bool     `json:"predicted"`
	VehicleID                  string   `json:"vehicleId"`
	Position                   Location `json:"position"`
	LastKnownLocation          Location `json:"lastKnownLocation"`
	Orientation                float64  `json:"orientation"`
	LastKnownOrientation       float64  `json:"lastKnownOrientation"`
	ScheduleDeviation          int      `json:"scheduleDeviation"`
	DistanceAlongTrip          float64  `json:"distanceAlongTrip"`
	ScheduledDistanceAlongTrip float64  `json:"scheduledDistanceAlongTrip"`
	TotalDistanceAlongTrip     float64  `json:"totalDistanceAlongTrip"`
	LastKnownDistanceAlongTrip float64  `json:"lastKnownDistanceAlongTrip"`
	LastUpdateTime             int64    `json:"lastUpdateTime"`
	LastLocationUpdateTime     int64    `json:"lastLocationUpdateTime"`
	BlockTripSequence          int      `json:"blockTripSequence"`
	ClosestStop                string   `json:"closestStop"`
	ClosestStopTimeOffset      int      `json:"closestStopTimeOffset"`
	NextStop                   string   `json:"nextStop"`
	NextStopTimeOffset         int      `json:"nextStopTimeOffset"`
	OccupancyStatus            string   `json:"occupancyStatus"`
	OccupancyCount             int      `json:"occupancyCount"`
	OccupancyCapacity          int      `json:"occupancyCapacity"`
	SituationIDs               []string `json:"situationIds"`
	Scheduled                  bool     `json:"scheduled"`
}
