package gtfs

import (
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
