package models

type TripDetails struct {
	TripID       string      `json:"tripId"`
	ServiceDate  int64       `json:"serviceDate"`
	Frequency    *Frequency  `json:"frequency,omitempty"`
	Status       *TripStatus `json:"status,omitempty"`
	Schedule     *Schedule   `json:"schedule"`
	SituationIDs []string    `json:"situationIds,omitempty"`
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
