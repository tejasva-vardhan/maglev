package models

type Edge struct {
	A CoordinatePoint
	B CoordinatePoint
}
type CoordinatePoint struct {
	Lat float64
	Lon float64
}

func NewEdge(a, b CoordinatePoint) Edge {
	if ComparePoints(a, b) <= 0 {
		return Edge{A: a, B: b}
	}
	return Edge{A: b, B: a}
}

func ComparePoints(a, b CoordinatePoint) int {
	if a.Lat < b.Lat {
		return -1
	}
	if a.Lat > b.Lat {
		return 1
	}
	if a.Lon < b.Lon {
		return -1
	}
	if a.Lon > b.Lon {
		return 1
	}
	return 0
}
