package gtfs

import (
	"github.com/jamespfennell/gtfs"
	"github.com/stretchr/testify/assert"
	"maglev.onebusaway.org/internal/models"
	"testing"
)

func TestManager_GetAgencies(t *testing.T) {
	testCases := []struct {
		name     string
		dataPath string
	}{
		{
			name:     "FromLocalFile",
			dataPath: models.GetFixturePath(t, "gtfs.zip"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gtfsConfig := Config{
				GtfsURL: tc.dataPath,
			}
			manager, err := InitGTFSManager(gtfsConfig)
			assert.Nil(t, err)

			agencies := manager.GetAgencies()
			assert.Equal(t, 1, len(agencies))

			agency := agencies[0]
			assert.Equal(t, "40", agency.Id)
			assert.Equal(t, "Sound Transit", agency.Name)
			assert.Equal(t, "https://www.soundtransit.org", agency.Url)
			assert.Equal(t, "America/Los_Angeles", agency.Timezone)
			assert.Equal(t, "en", agency.Language)
			assert.Equal(t, "1-888-889-6368", agency.Phone)
			assert.Equal(t, "https://www.soundtransit.org/ride-with-us/how-to-pay/fares", agency.FareUrl)
			assert.Equal(t, "main@soundtransit.org", agency.Email)
		})
	}
}

func TestGetRegionBounds(t *testing.T) {
	tests := []struct {
		name            string
		shapes          []gtfs.Shape
		expectedLat     float64
		expectedLon     float64
		expectedLatSpan float64
		expectedLonSpan float64
	}{
		{
			name: "Single Shape With Multiple Points",
			shapes: []gtfs.Shape{
				{
					ID: "shape1",
					Points: []gtfs.ShapePoint{
						{Latitude: 47.0, Longitude: -122.0},
						{Latitude: 48.0, Longitude: -121.0},
					},
				},
			},
			expectedLat:     47.5,
			expectedLon:     -121.5,
			expectedLatSpan: 1.0,
			expectedLonSpan: 1.0,
		},
		{
			name: "Multiple Shapes",
			shapes: []gtfs.Shape{
				{
					ID: "shape1",
					Points: []gtfs.ShapePoint{
						{Latitude: 47.0, Longitude: -122.0},
						{Latitude: 48.0, Longitude: -121.0},
					},
				},
				{
					ID: "shape2",
					Points: []gtfs.ShapePoint{
						{Latitude: 46.5, Longitude: -123.0},
						{Latitude: 48.5, Longitude: -120.5},
					},
				},
			},
			expectedLat:     47.5,
			expectedLon:     -121.75,
			expectedLatSpan: 2.0,
			expectedLonSpan: 2.5,
		},
		{
			name:            "No Shapes",
			shapes:          []gtfs.Shape{},
			expectedLat:     0.0,
			expectedLon:     0.0,
			expectedLatSpan: 0.0,
			expectedLonSpan: 0.0,
		},
		{
			name: "Shape With No Points",
			shapes: []gtfs.Shape{
				{
					ID:     "empty",
					Points: []gtfs.ShapePoint{},
				},
			},
			expectedLat:     0.0,
			expectedLon:     0.0,
			expectedLatSpan: 0.0,
			expectedLonSpan: 0.0,
		},
		{
			name: "Real Example",
			shapes: []gtfs.Shape{
				{
					ID: "shape1",
					Points: []gtfs.ShapePoint{
						{Latitude: 47.5665345, Longitude: -122.3032715},
						{Latitude: 47.6023246, Longitude: -122.3308378},
						{Latitude: 47.6534563, Longitude: -122.3472905},
						{Latitude: 47.6932255, Longitude: -122.3116085},
					},
				},
				{
					ID: "shape2",
					Points: []gtfs.ShapePoint{
						{Latitude: 47.5998745, Longitude: -122.3274812},
						{Latitude: 47.6137821, Longitude: -122.3450112},
						{Latitude: 47.6423451, Longitude: -122.3274812},
					},
				},
			},
			expectedLat:     47.629880000000,
			expectedLon:     -122.32528099999,
			expectedLatSpan: 0.12669099999999,
			expectedLonSpan: 0.04401900000000,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gtfsConfig := Config{
				GtfsURL: models.GetFixturePath(t, "gtfs.zip"),
			}
			manager, err := InitGTFSManager(gtfsConfig)

			// Set custom shapes
			manager.gtfsData.Shapes = tc.shapes
			if err != nil {
				t.Fatalf("Failed to initialize GTFS manager: %v", err)
			}
			lat, lon, latSpan, lonSpan := manager.GetRegionBounds()

			if tc.name == "No Shapes" || tc.name == "Shape With No Points" {
				t.Logf("Test case %s returned: lat=%f, lon=%f, latSpan=%f, lonSpan=%f",
					tc.name, lat, lon, latSpan, lonSpan)
			} else {
				assert.InDeltaf(t, tc.expectedLat, lat, 0.00000001, "Latitude mismatch in %s", tc.name)
				assert.InDeltaf(t, tc.expectedLon, lon, 0.00000001, "Longitude mismatch in %s", tc.name)
				assert.InDeltaf(t, tc.expectedLatSpan, latSpan, 0.00000001, "Latitude span mismatch in %s", tc.name)
				assert.InDeltaf(t, tc.expectedLonSpan, lonSpan, 0.00000001, "Longitude span mismatch in %s", tc.name)
			}
		})
	}
}
