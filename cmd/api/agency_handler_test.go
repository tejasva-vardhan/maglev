package main

import (
	"encoding/json"
	"maglev.onebusaway.org/internal/gtfs"
	"maglev.onebusaway.org/internal/models"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgencyHandlerReturnsAgencyWhenItExists(t *testing.T) {
	gtfsConfig := gtfs.Config{
		GtfsURL: filepath.Join("../../testdata", "gtfs.zip"),
	}
	gtfsManager, err := gtfs.InitGTFSManager(gtfsConfig)
	require.NoError(t, err)

	agencies := gtfsManager.GetAgencies()
	require.NotEmpty(t, agencies)
	agencyID := agencies[0].Id

	app := &application{
		config: config{
			env:     "test",
			apiKeys: []string{"TEST"},
		},
		gtfsManager: gtfsManager,
	}

	server := httptest.NewServer(app.routes())
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/where/agency/" + agencyID + ".json?key=TEST")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response models.ResponseModel
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}()

	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "OK", response.Text)

	data, ok := response.Data.(map[string]interface{})
	require.True(t, ok)

	entry, ok := data["entry"].(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, agencies[0].Id, entry["id"])
	assert.Equal(t, agencies[0].Name, entry["name"])
	assert.Equal(t, agencies[0].Url, entry["url"])
	assert.Equal(t, agencies[0].Timezone, entry["timezone"])
}

func TestAgencyHandlerReturnsNullWhenAgencyDoesNotExist(t *testing.T) {
	gtfsConfig := gtfs.Config{
		GtfsURL: filepath.Join("../../testdata", "gtfs.zip"),
	}
	gtfsManager, err := gtfs.InitGTFSManager(gtfsConfig)
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

	resp, err := http.Get(server.URL + "/api/where/agency/non-existent-id.json?key=TEST")
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var response models.ResponseModel
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}()

	assert.Equal(t, 404, response.Code)
	assert.Equal(t, "resource not found", response.Text)
	assert.Nil(t, response.Data)
}

func TestAgencyHandlerRequiresValidApiKey(t *testing.T) {
	gtfsConfig := gtfs.Config{
		GtfsURL: filepath.Join("../../testdata", "gtfs.zip"),
	}
	gtfsManager, err := gtfs.InitGTFSManager(gtfsConfig)
	require.NoError(t, err)

	agencies := gtfsManager.GetAgencies()
	require.NotEmpty(t, agencies)
	agencyID := agencies[0].Id

	app := &application{
		config: config{
			env:     "test",
			apiKeys: []string{"TEST"},
		},
		gtfsManager: gtfsManager,
	}

	server := httptest.NewServer(app.routes())
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/where/agency/" + agencyID + ".json?key=INVALID")
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}()
}
