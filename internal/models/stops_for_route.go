package models

type Polyline struct {
	Length int    `json:"length"`
	Levels string `json:"levels"`
	Points string `json:"points"`
}

type StopGroupName struct {
	Name  string   `json:"name"`
	Names []string `json:"names"`
	Type  string   `json:"type"`
}

type StopGroup struct {
	ID        string        `json:"id"`
	Name      StopGroupName `json:"name"`
	StopIds   []string      `json:"stopIds"`
	Polylines []Polyline    `json:"polylines"`
}

type StopGrouping struct {
	Type       string      `json:"type"`
	Ordered    bool        `json:"ordered"`
	StopGroups []StopGroup `json:"stopGroups"`
}

type RouteEntry struct {
	RouteID       string         `json:"routeId"`
	StopIds       []string       `json:"stopIds"`
	StopGroupings []StopGrouping `json:"stopGroupings"`
	Polylines     []Polyline     `json:"polylines"`
}
