package models

// Shape represents a raw shape point from the GTFS database
type Shape struct {
	ID              int64   `json:"id"`
	ShapeID         string  `json:"shape_id"`
	Lat             float64 `json:"lat"`
	Lon             float64 `json:"lon"`
	ShapePtSequence int64   `json:"shape_pt_sequence"`
}

// ShapeEntry represents a shape entry for the API response
type ShapeEntry struct {
	Points string `json:"points"`
	Length int    `json:"length"`
	Levels string `json:"levels"`
}
