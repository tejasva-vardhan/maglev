package gtfs

import "github.com/OneBusAway/go-gtfs"

// ComputeRegionBounds calculates the geographic boundaries of the GTFS region
// from all shape points. Returns nil if no shapes exist.
func ComputeRegionBounds(shapes []gtfs.Shape) *RegionBounds {
	if len(shapes) == 0 {
		return nil
	}

	var minLat, maxLat, minLon, maxLon float64
	first := true

	for _, shape := range shapes {
		for _, point := range shape.Points {
			if first {
				minLat = point.Latitude
				maxLat = point.Latitude
				minLon = point.Longitude
				maxLon = point.Longitude
				first = false
				continue
			}

			if point.Latitude < minLat {
				minLat = point.Latitude
			}
			if point.Latitude > maxLat {
				maxLat = point.Latitude
			}
			if point.Longitude < minLon {
				minLon = point.Longitude
			}
			if point.Longitude > maxLon {
				maxLon = point.Longitude
			}
		}
	}

	return &RegionBounds{
		Lat:     (minLat + maxLat) / 2,
		Lon:     (minLon + maxLon) / 2,
		LatSpan: maxLat - minLat,
		LonSpan: maxLon - minLon,
	}
}

// IMPORTANT: Caller must hold manager.RLock() before calling this method.
func (manager *Manager) GetRegionBounds() (lat, lon, latSpan, lonSpan float64) {
	if manager.regionBounds == nil {
		return 0, 0, 0, 0
	}
	return manager.regionBounds.Lat, manager.regionBounds.Lon, manager.regionBounds.LatSpan, manager.regionBounds.LonSpan
}
