package models

type Frequency struct {
	StartTime int64 `json:"startTime"`
	EndTime   int64 `json:"endTime"`
	Headway   int   `json:"headway"`
}
