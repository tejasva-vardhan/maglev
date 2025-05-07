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

func TestRoutesForAgencyHandlerEndToEnd(t *testing.T) {
	gtfsPath := filepath.Join("../../testdata", "gtfs.zip")
	gtfsManager, err := gtfs.InitGTFSManager(gtfsPath)
	require.NoError(t, err)

	agencies := gtfsManager.GetAgencies()
	require.NotEmpty(t, agencies)
	agencyId := agencies[0].Id

	app := &application{
		config: config{
			env:     "test",
			apiKeys: []string{"TEST"},
		},
		gtfsManager: gtfsManager,
	}

	server := httptest.NewServer(app.routes())
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/where/routes-for-agency/" + agencyId + ".json?key=TEST")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response models.ResponseModel
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	defer resp.Body.Close() // nolint:errcheck

	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "OK", response.Text)

	data, ok := response.Data.(map[string]interface{})
	require.True(t, ok)

	// Check that we have a list of routes
	_, ok = data["list"].([]interface{})
	require.True(t, ok)

	refs, ok := data["references"].(map[string]interface{})
	require.True(t, ok)

	refAgencies, ok := refs["agencies"].([]interface{})
	require.True(t, ok)
	assert.Len(t, refAgencies, 1)
}
