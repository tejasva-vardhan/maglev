package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgenciesWithCoverageHandlerRequiresValidApiKey(t *testing.T) {
	_, resp, model := serveAndRetrieveEndpoint(t, "/api/where/agencies-with-coverage.json?key=invalid")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, http.StatusUnauthorized, model.Code)
	assert.Equal(t, "permission denied", model.Text)
}

func TestAgenciesWithCoverageHandlerEndToEnd(t *testing.T) {
	_, resp, model := serveAndRetrieveEndpoint(t, "/api/where/agencies-with-coverage.json?key=TEST")

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, http.StatusOK, model.Code)
	assert.Equal(t, "OK", model.Text)

	data, ok := model.Data.(map[string]interface{})
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
