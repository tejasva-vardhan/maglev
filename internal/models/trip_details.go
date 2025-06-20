package models

type TripDetails struct {
	TripID       string      `json:"tripId"`
	ServiceDate  int64       `json:"serviceDate"`
	Frequency    *Frequency  `json:"frequency"`
	Status       *TripStatus `json:"status"`
	Schedule     *Schedule   `json:"schedule"`
	SituationIDs []string    `json:"situationIds"`
}

func NewTripDetails(trip Trip, tripID string, serviceDate int64, frequency *Frequency, status *TripStatus, schedule *Schedule, situationIDs []string) *TripDetails {
	return &TripDetails{
		TripID:       tripID,
		ServiceDate:  serviceDate,
		Frequency:    frequency,
		Status:       status,
		Schedule:     schedule,
		SituationIDs: situationIDs,
	}

}

func NewEmptyTripDetails() *TripDetails {
	return &TripDetails{
		TripID:       "",
		ServiceDate:  0,
		Frequency:    nil,
		Status:       nil,
		Schedule:     nil,
		SituationIDs: []string{},
	}
}
