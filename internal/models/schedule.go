package models

type Schedule struct {
	Frequency      int64      `json:"frequency"`
	NextTripID     string     `json:"nextTripId"`
	PreviousTripID string     `json:"previousTripId"`
	StopTimes      []StopTime `json:"stopTimes"`
	TimeZone       string     `json:"timeZone"`
}

func NewSchedule(frequency int64, nextTripID, previousTripID string, stopTimes []StopTime, timeZone string) *Schedule {
	return &Schedule{
		Frequency:      frequency,
		NextTripID:     nextTripID,
		PreviousTripID: previousTripID,
		StopTimes:      stopTimes,
		TimeZone:       timeZone,
	}
}
