package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"maglev.onebusaway.org/internal/gtfs"
	"maglev.onebusaway.org/internal/models"
)

func TestAgenciesWithCoverageHandlerEndToEnd(t *testing.T) {
	gtfsPath := filepath.Join("../../testdata", "gtfs.zip")
	gtfsManager, err := gtfs.InitGTFSManager(gtfsPath)
	require.NoError(t, err)

	app := &application{
		config: config{
			env:     "test",
			apiKeys: []string{"TEST"},
		},
		gtfsManager: gtfsManager,
	}

	server := httptest.NewServer(app.routes())
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/where/agencies-with-coverage.json?key=TEST")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response models.ResponseModel
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "OK", response.Text)

	data, ok := response.Data.(map[string]interface{})
	require.True(t, ok)

	list, ok := data["list"].([]interface{})
	require.True(t, ok)
	assert.Len(t, list, 1)

	agencyCoverage, ok := list[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "40", agencyCoverage["agencyId"])
	assert.InDelta(t, 47.5665345, agencyCoverage["lat"], 1e-8)
	assert.InDelta(t, 0.826691000000004, agencyCoverage["latSpan"], 1e-8)
	assert.InDelta(t, -122.31623250000001, agencyCoverage["lon"], 1e-8)
	assert.InDelta(t, 0.36574099999999987, agencyCoverage["lonSpan"], 1e-8)

	refs, ok := data["references"].(map[string]interface{})
	require.True(t, ok)

	refAgencies, ok := refs["agencies"].([]interface{})
	require.True(t, ok)
	assert.Len(t, refAgencies, 1)

	agencyRef, ok := refAgencies[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "40", agencyRef["id"])
	assert.Equal(t, "Sound Transit", agencyRef["name"])
	assert.Equal(t, "https://www.soundtransit.org", agencyRef["url"])
	assert.Equal(t, "America/Los_Angeles", agencyRef["timezone"])
	assert.Equal(t, "en", agencyRef["lang"])
	assert.Equal(t, "1-888-889-6368", agencyRef["phone"])
	assert.Equal(t, "main@soundtransit.org", agencyRef["email"])
	assert.Equal(t, "https://www.soundtransit.org/ride-with-us/how-to-pay/fares", agencyRef["fareUrl"])
	assert.Equal(t, "", agencyRef["disclaimer"])
	assert.False(t, agencyRef["privateService"].(bool))
	// Ensure no extra fields
	assert.Len(t, agencyRef, 10)

	assert.Empty(t, refs["routes"])
	assert.Empty(t, refs["situations"])
	assert.Empty(t, refs["stopTimes"])
	assert.Empty(t, refs["stops"])
	assert.Empty(t, refs["trips"])
}
