package restapi

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"log/slog"
	"maglev.onebusaway.org/internal/app"
	"maglev.onebusaway.org/internal/appconf"
	"maglev.onebusaway.org/internal/gtfs"
	"maglev.onebusaway.org/internal/logging"
	"maglev.onebusaway.org/internal/models"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

// createTestApi creates a new restAPI instance with a GTFS manager initialized for use in tests.
func createTestApi(t *testing.T) *RestAPI {
	gtfsConfig := gtfs.Config{
		GtfsURL:      filepath.Join("../../testdata", "raba.zip"),
		GTFSDataPath: ":memory:",
	}
	gtfsManager, err := gtfs.InitGTFSManager(gtfsConfig)
	require.NoError(t, err)

	app := &app.Application{
		Config: appconf.Config{
			Env:     appconf.EnvFlagToEnvironment("test"),
			ApiKeys: []string{"TEST"},
		},
		GtfsConfig:  gtfsConfig,
		GtfsManager: gtfsManager,
	}

	api := &RestAPI{Application: app}

	return api
}

// serveAndRetrieveEndpoint sets up a test server, makes a request to the specified endpoint, and returns the response
// and decoded model.
func serveAndRetrieveEndpoint(t *testing.T, endpoint string) (*RestAPI, *http.Response, models.ResponseModel) {
	api := createTestApi(t)
	resp, model := serveApiAndRetrieveEndpoint(t, api, endpoint)
	return api, resp, model
}

func serveApiAndRetrieveEndpoint(t *testing.T, api *RestAPI, endpoint string) (*http.Response, models.ResponseModel) {
	mux := http.NewServeMux()
	api.SetRoutes(mux)
	server := httptest.NewServer(mux)
	defer server.Close()
	resp, err := http.Get(server.URL + endpoint)
	require.NoError(t, err)
	defer logging.SafeCloseWithLogging(resp.Body,
		slog.Default().With(slog.String("component", "test")),
		"http_response_body")

	var response models.ResponseModel
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	return resp, response
}
