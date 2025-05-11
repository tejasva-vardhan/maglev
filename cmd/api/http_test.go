package main

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"maglev.onebusaway.org/internal/app"
	"maglev.onebusaway.org/internal/gtfs"
	"maglev.onebusaway.org/internal/models"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

// createTestApi creates a new restAPI instance with a GTFS manager initialized for use in tests.
func createTestApi(t *testing.T) *restAPI {
	gtfsConfig := gtfs.Config{
		GtfsURL: filepath.Join("../../testdata", "gtfs.zip"),
	}
	gtfsManager, err := gtfs.InitGTFSManager(gtfsConfig)
	require.NoError(t, err)

	app := &app.Application{
		Config: app.Config{
			Env:     "test",
			ApiKeys: []string{"TEST"},
		},
		GtfsManager: gtfsManager,
	}

	api := &restAPI{app: app}

	return api
}

// serveAndRetrieveEndpoint sets up a test server, makes a request to the specified endpoint, and returns the response
// and decoded model.
func serveAndRetrieveEndpoint(t *testing.T, endpoint string) (*restAPI, *http.Response, models.ResponseModel) {
	api := createTestApi(t)
	resp, model := serveApiAndRetrieveEndpoint(t, api, endpoint)
	return api, resp, model
}

func serveApiAndRetrieveEndpoint(t *testing.T, api *restAPI, endpoint string) (*http.Response, models.ResponseModel) {
	server := httptest.NewServer(api.routes())
	defer server.Close()
	resp, err := http.Get(server.URL + endpoint)
	require.NoError(t, err)
	defer resp.Body.Close() // nolint:errcheck

	var response models.ResponseModel
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	return resp, response
}
