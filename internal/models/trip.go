package models

type Trip struct {
	BlockID        string `json:"blockId"`
	DirectionID    int64  `json:"directionId"`
	ID             string `json:"id"`
	RouteID        string `json:"routeId"`
	ServiceID      string `json:"serviceId"`
	ShapeID        string `json:"shapeId"`
	TripHeadsign   string `json:"tripHeadsign"`
	TripShortName  string `json:"tripShortName"`
	RouteShortName string `json:"routeShortName"`
	PeakOffPeak    int64  `json:"peakOffPeak"`
	TimeZone       string `json:"timeZone"`
}

type TripResponse struct {
	*Trip
}

func NewTripResponse(trip *Trip, routeShortName, timeZone string, peakOffPeak int) *TripResponse {
	return &TripResponse{
		Trip: trip,
	}
}
