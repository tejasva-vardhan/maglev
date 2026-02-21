package gtfs

import "github.com/OneBusAway/go-gtfs"

// ComputeRegionBounds calculates the geographic boundaries of the GTFS region
// from all shape points, falling back to stops if no shapes exist.
func ComputeRegionBounds(shapes []gtfs.Shape, stops []gtfs.Stop) *RegionBounds {
	if len(shapes) == 0 && len(stops) == 0 {
		return nil
	}

	var minLat, maxLat, minLon, maxLon float64
	first := true

	if len(shapes) > 0 {
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
	} else {
		for _, stop := range stops {
			if stop.Latitude == nil || stop.Longitude == nil {
				continue
			}
			if first {
				minLat = *stop.Latitude
				maxLat = *stop.Latitude
				minLon = *stop.Longitude
				maxLon = *stop.Longitude
				first = false
				continue
			}

			if *stop.Latitude < minLat {
				minLat = *stop.Latitude
			}
			if *stop.Latitude > maxLat {
				maxLat = *stop.Latitude
			}
			if *stop.Longitude < minLon {
				minLon = *stop.Longitude
			}
			if *stop.Longitude > maxLon {
				maxLon = *stop.Longitude
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
