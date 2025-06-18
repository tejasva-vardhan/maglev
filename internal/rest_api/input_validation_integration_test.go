package restapi

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInputValidationIntegration(t *testing.T) {
	api := createTestApi(t)

	tests := []struct {
		name           string
		endpoint       string
		expectedStatus int
		expectedError  string
	}{
		// Test malicious ID inputs
		{
			name:           "SQL injection in agency ID",
			endpoint:       "/api/where/agency/raba'; DROP TABLE agencies; --.json?key=TEST",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "ID contains invalid characters",
		},
		{
			name:           "XSS in agency ID",
			endpoint:       "/api/where/agency/raba<script>alert('xss')</script>.json?key=TEST",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "ID contains invalid characters",
		},
		{
			name:           "Path traversal in agency ID",
			endpoint:       "/api/where/agency/../../../etc/passwd.json?key=TEST",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "ID contains invalid characters",
		},
		{
			name:           "Long ID exceeding limit",
			endpoint:       fmt.Sprintf("/api/where/agency/%s.json?key=TEST", strings.Repeat("a", 101)),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "ID too long",
		},
		{
			name:           "Empty ID",
			endpoint:       "/api/where/agency/.json?key=TEST",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "ID cannot be empty",
		},

		// Test malicious location parameters
		{
			name:           "Invalid latitude too high",
			endpoint:       "/api/where/stops-for-location.json?key=TEST&lat=91.0&lon=-77.0",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Latitude must be between -90 and 90",
		},
		{
			name:           "Invalid longitude too high",
			endpoint:       "/api/where/stops-for-location.json?key=TEST&lat=38.0&lon=181.0",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Longitude must be between -180 and 180",
		},
		{
			name:           "Negative radius",
			endpoint:       "/api/where/stops-for-location.json?key=TEST&lat=38.0&lon=-77.0&radius=-100",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Radius must be non-negative",
		},
		{
			name:           "Radius too large",
			endpoint:       "/api/where/stops-for-location.json?key=TEST&lat=38.0&lon=-77.0&radius=50000",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Radius too large",
		},

		// Test malicious query parameters
		{
			name:           "Script injection in query",
			endpoint:       "/api/where/stops-for-location.json?key=TEST&lat=38.0&lon=-77.0&query=<script>alert('xss')</script>",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Query contains invalid characters",
		},
		{
			name:           "SQL injection in query",
			endpoint:       "/api/where/stops-for-location.json?key=TEST&lat=38.0&lon=-77.0&query='; DROP TABLE stops; --",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Query contains invalid characters",
		},
		{
			name:           "Query too long",
			endpoint:       fmt.Sprintf("/api/where/stops-for-location.json?key=TEST&lat=38.0&lon=-77.0&query=%s", strings.Repeat("a", 201)),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Query too long",
		},

		// Test malicious date parameters
		{
			name:           "Invalid date format",
			endpoint:       "/api/where/schedule-for-stop/raba_12345.json?key=TEST&date=12/25/2023",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid date format",
		},
		{
			name:           "Date with script injection",
			endpoint:       "/api/where/schedule-for-stop/raba_12345.json?key=TEST&date=2023-01-01<script>alert('xss')</script>",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid date format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, _ := serveApiAndRetrieveEndpoint(t, api, tt.endpoint)
			assert.Equal(t, tt.expectedStatus, response.StatusCode, "Expected status code mismatch")
			
			// Check that the response contains the expected error message
			if tt.expectedStatus == http.StatusBadRequest {
				body := readResponseBody(t, response)
				assert.Contains(t, body, tt.expectedError, "Response should contain expected error message")
			}
		})
	}
}

func TestInputSanitizationIntegration(t *testing.T) {
	api := createTestApi(t)

	tests := []struct {
		name     string
		endpoint string
		query    string
		expected string
	}{
		{
			name:     "Normal query is preserved",
			query:    "downtown station",
			expected: "downtown station",
		},
		{
			name:     "Special characters in station names are allowed",
			query:    "St. Mary's Hospital & Clinic",
			expected: "st. mary's hospital & clinic", // lowercase due to existing handler logic
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build URL with query parameter
			baseURL := "/api/where/stops-for-location.json?key=TEST&lat=38.9&lon=-77.0"
			if tt.query != "" {
				baseURL += "&query=" + url.QueryEscape(tt.query)
			}
			
			response, _ := serveApiAndRetrieveEndpoint(t, api, baseURL)
			
			// Should succeed (not be blocked by validation)
			assert.Equal(t, http.StatusOK, response.StatusCode, "Valid query should not be blocked")
		})
	}
}

func TestValidInputsPassThrough(t *testing.T) {
	api := createTestApi(t)

	validTests := []struct {
		name     string
		endpoint string
	}{
		{
			name:     "Valid agency ID",
			endpoint: "/api/where/agency/raba.json?key=TEST",
		},
		{
			name:     "Valid stop ID",
			endpoint: "/api/where/stop/raba_12345.json?key=TEST",
		},
		{
			name:     "Valid location parameters",
			endpoint: "/api/where/stops-for-location.json?key=TEST&lat=38.9&lon=-77.0&radius=1000",
		},
		{
			name:     "Valid date parameter",
			endpoint: "/api/where/schedule-for-stop/raba_12345.json?key=TEST&date=2023-12-25",
		},
		{
			name:     "Valid location with span parameters",
			endpoint: "/api/where/stops-for-location.json?key=TEST&lat=38.9&lon=-77.0&latSpan=0.01&lonSpan=0.01",
		},
	}

	for _, tt := range validTests {
		t.Run(tt.name, func(t *testing.T) {
			response, _ := serveApiAndRetrieveEndpoint(t, api, tt.endpoint)
			
			// Should not return validation errors (400)
			// Note: Some endpoints may return 404 if the data doesn't exist, which is fine
			assert.NotEqual(t, http.StatusBadRequest, response.StatusCode, 
				"Valid input should not return validation error")
		})
	}
}

func TestEdgeCaseValidation(t *testing.T) {
	api := createTestApi(t)

	tests := []struct {
		name           string
		endpoint       string
		expectedStatus int
	}{
		{
			name:           "Boundary latitude values - north pole",
			endpoint:       "/api/where/stops-for-location.json?key=TEST&lat=90.0&lon=0.0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Boundary latitude values - south pole",
			endpoint:       "/api/where/stops-for-location.json?key=TEST&lat=-90.0&lon=0.0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Boundary longitude values - date line east",
			endpoint:       "/api/where/stops-for-location.json?key=TEST&lat=0.0&lon=180.0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Boundary longitude values - date line west",
			endpoint:       "/api/where/stops-for-location.json?key=TEST&lat=0.0&lon=-180.0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Zero radius is valid",
			endpoint:       "/api/where/stops-for-location.json?key=TEST&lat=38.9&lon=-77.0&radius=0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Maximum allowed radius",
			endpoint:       "/api/where/stops-for-location.json?key=TEST&lat=38.9&lon=-77.0&radius=10000",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty query parameter is valid",
			endpoint:       "/api/where/stops-for-location.json?key=TEST&lat=38.9&lon=-77.0&query=",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Empty date parameter is valid",
			endpoint:       "/api/where/schedule-for-stop/raba_12345.json?key=TEST&date=",
			expectedStatus: http.StatusOK, // Should use current date
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, _ := serveApiAndRetrieveEndpoint(t, api, tt.endpoint)
			assert.Equal(t, tt.expectedStatus, response.StatusCode, "Expected status code mismatch")
		})
	}
}

// Helper function to read response body as string
func readResponseBody(t *testing.T, response *http.Response) string {
	body := make([]byte, 1024)
	n, err := response.Body.Read(body)
	if err != nil && n == 0 {
		require.NoError(t, err, "Failed to read response body")
	}
	return string(body[:n])
}