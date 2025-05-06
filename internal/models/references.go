package models

// ReferencesModel References model for related data
type ReferencesModel struct {
	Agencies   []AgencyReference `json:"agencies"`
	Routes     []interface{}     `json:"routes"`
	Situations []interface{}     `json:"situations"`
	StopTimes  []interface{}     `json:"stopTimes"`
	Stops      []interface{}     `json:"stops"`
	Trips      []interface{}     `json:"trips"`
}

// NewEmptyReferences creates a new empty References model with initialized empty slices
func NewEmptyReferences() ReferencesModel {
	return ReferencesModel{
		Agencies:   []AgencyReference{},
		Routes:     []interface{}{},
		Situations: []interface{}{},
		StopTimes:  []interface{}{},
		Stops:      []interface{}{},
		Trips:      []interface{}{},
	}
}
