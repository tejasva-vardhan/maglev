package main

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"maglev.onebusaway.org/internal/gtfs"
	"maglev.onebusaway.org/internal/models"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

// createTestApp creates a new application instance with a GTFS manager initialized for use in tests.
func createTestApp(t *testing.T) *application {
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

	return app
}

// serveAndRetrieveEndpoint sets up a test server, makes a request to the specified endpoint, and returns the response
// and decoded model.
func serveAndRetrieveEndpoint(t *testing.T, endpoint string) (*application, *http.Response, models.ResponseModel) {
	app := createTestApp(t)
	resp, model := serveAppAndRetrieveEndpoint(t, app, endpoint)
	return app, resp, model
}

func serveAppAndRetrieveEndpoint(t *testing.T, app *application, endpoint string) (*http.Response, models.ResponseModel) {
	server := httptest.NewServer(app.routes())
	defer server.Close()
	resp, err := http.Get(server.URL + endpoint)
	require.NoError(t, err)
	defer resp.Body.Close() // nolint:errcheck

	var response models.ResponseModel
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	return resp, response
}
