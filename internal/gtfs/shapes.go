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

	updateBounds := func(lat, lon float64) {
		if first {
			minLat = lat
			maxLat = lat
			minLon = lon
			maxLon = lon
			first = false
			return
		}
		if lat < minLat {
			minLat = lat
		}
		if lat > maxLat {
			maxLat = lat
		}
		if lon < minLon {
			minLon = lon
		}
		if lon > maxLon {
			maxLon = lon
		}
	}

	if len(shapes) > 0 {
		for _, shape := range shapes {
			for _, point := range shape.Points {
				updateBounds(point.Latitude, point.Longitude)
			}
		}
	} else {
		for _, stop := range stops {
			if stop.Latitude == nil || stop.Longitude == nil {
				continue
			}
			updateBounds(*stop.Latitude, *stop.Longitude)
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
