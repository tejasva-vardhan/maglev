package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"maglev.onebusaway.org/internal/gtfs"
	"maglev.onebusaway.org/internal/models"
)

func TestRoutesForAgencyHandler(t *testing.T) {
	gtfsPath := filepath.Join("../../testdata", "gtfs.zip")
	gtfsManager, err := gtfs.InitGTFSManager(gtfsPath)
	require.NoError(t, err, "Failed to initialize GTFS manager with test data")

	gtfsManager.PrintStatistics()

	agencies := gtfsManager.GetAgencies()
	require.NotEmpty(t, agencies, "No agencies found in test GTFS data")

	agencyId := agencies[0].Id

	app := &application{
		config: config{
			env: "test",
		},
		gtfsManager: gtfsManager,
	}

	tests := []struct {
		name           string
		agencyID       string
		expectedStatus int
		validateBody   func(*testing.T, []byte)
	}{
		{
			name:           "Valid agency",
			agencyID:       agencyId,
			expectedStatus: http.StatusOK,
			validateBody: func(t *testing.T, body []byte) {
				var response models.ResponseModel
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)

				assert.Equal(t, 200, response.Code)
				assert.Equal(t, "OK", response.Text)
				assert.Equal(t, 2, response.Version)

				data, ok := response.Data.(map[string]interface{})
				require.True(t, ok)

				_, ok = data["list"].([]interface{})
				require.True(t, ok)

				refs, ok := data["references"].(map[string]interface{})
				require.True(t, ok)

				agencies, ok := refs["agencies"].([]interface{})
				require.True(t, ok)
				assert.Len(t, agencies, 1)

				agency, ok := agencies[0].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, agencyId, agency["id"])
			},
		},
		{
			name:           "Invalid agency",
			agencyID:       "nonexistent-agency-id",
			expectedStatus: http.StatusNotFound,
			validateBody: func(t *testing.T, body []byte) {
				assert.Equal(t, "null\n", string(body))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/api/where/routes-for-agency/"+tt.agencyID+".json", nil)
			require.NoError(t, err)

			params := httprouter.Params{
				{Key: "id.json", Value: tt.agencyID + ".json"},
			}
			req = req.WithContext(context.WithValue(req.Context(), httprouter.ParamsKey, params))

			rr := httptest.NewRecorder()

			app.routesForAgencyHandler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.validateBody != nil {
				tt.validateBody(t, rr.Body.Bytes())
			}
		})
	}
}

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

	assert.Equal(t, 200, response.Code)
	assert.Equal(t, "OK", response.Text)

	data, ok := response.Data.(map[string]interface{})
	require.True(t, ok)

	refs, ok := data["references"].(map[string]interface{})
	require.True(t, ok)

	refAgencies, ok := refs["agencies"].([]interface{})
	require.True(t, ok)
	assert.Len(t, refAgencies, 1)
}
