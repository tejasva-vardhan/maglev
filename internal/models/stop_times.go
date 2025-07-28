package models

type StopTimes struct {
	StopTimes []StopTime `json:"stop_times"`
}

type StopTime struct {
	ArrivalTime         int     `json:"arrivalTime"`
	DepartureTime       int     `json:"departureTime"`
	DropOffType         int     `json:"dropOffType,omitempty"`
	PickupType          int     `json:"pickupType,omitempty"`
	StopID              string  `json:"stopId"`
	StopHeadsign        string  `json:"stopHeadsign,omitempty"`
	DistanceAlongTrip   float64 `json:"distanceAlongTrip,omitempty"`
	HistoricalOccupancy string  `json:"historicalOccupancy,omitempty"`
}

func NewStopTime(arrivalTime, departureTime int, stopID, stopHeadsign string, distanceAlongTrip float64, historicalOccupancy string) StopTime {
	return StopTime{
		ArrivalTime:         arrivalTime,
		DepartureTime:       departureTime,
		StopID:              stopID,
		StopHeadsign:        stopHeadsign,
		DistanceAlongTrip:   distanceAlongTrip,
		HistoricalOccupancy: historicalOccupancy,
	}
}

func NewStopTimes(stopTimes []StopTime) StopTimes {
	return StopTimes{
		StopTimes: stopTimes,
	}
}
