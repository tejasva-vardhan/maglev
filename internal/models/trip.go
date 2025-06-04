package models

type Trip struct {
	BlockID       string `json:"blockId"`
	DirectionID   int64  `json:"directionId"`
	ID            string `json:"id"`
	RouteID       string `json:"routeId"`
	ServiceID     string `json:"serviceId"`
	ShapeID       string `json:"shapeId"`
	TripHeadsign  string `json:"tripHeadsign"`
	TripShortName string `json:"tripShortName"`
}

type TripResponse struct {
	*Trip

	RouteShortName string `json:"routeShortName"`
	TimeZone       string `json:"timeZone"`
	PeakOffPeak    int64  `json:"peakOffPeak"`
}

func NewTripResponse(trip *Trip, routeShortName, timeZone string, peakOffPeak int) *TripResponse {
	return &TripResponse{
		Trip:           trip,
		RouteShortName: routeShortName,
		TimeZone:       timeZone,
		PeakOffPeak:    0,
	}
}
