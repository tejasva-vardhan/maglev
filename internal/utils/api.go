package utils

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/jamespfennell/gtfs"
)

// ExtractCodeID extracts the `code_id` from a string in the format `{agency_id}_{code_id}`.
func ExtractCodeID(combinedID string) (string, error) {
	parts := strings.SplitN(combinedID, "_", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid format: %s", combinedID)
	}
	return parts[1], nil
}

// ExtractAgencyID extracts the `agency_id` from a string in the format `{agency_id}_{code_id}`.
func ExtractAgencyID(combinedID string) (string, error) {
	parts := strings.SplitN(combinedID, "_", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid format: %s", combinedID)
	}
	return parts[0], nil
}

// ExtractAgencyIDAndCodeID Extract AgencyIDAndCodeID extracts both `agency_id` and `code_id` from a string in the format `{agency_id}_{code_id}`.
func ExtractAgencyIDAndCodeID(combinedID string) (string, string, error) {
	parts := strings.SplitN(combinedID, "_", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid format: %s", combinedID)
	}
	return parts[0], parts[1], nil
}

// FormCombinedID forms a combined ID in the format `{agency_id}_{code_id}` using the given `agencyID` and `codeID`.
func FormCombinedID(agencyID, codeID string) string {
	if codeID == "" || agencyID == "" {
		return ""
	}
	return fmt.Sprintf("%s_%s", agencyID, codeID)
}

// MapWheelchairBoarding converts GTFS wheelchair boarding values to our API format
func MapWheelchairBoarding(wheelchairBoarding gtfs.WheelchairBoarding) string {
	switch wheelchairBoarding {
	case gtfs.WheelchairBoarding_Possible:
		return "ACCESSIBLE"
	case gtfs.WheelchairBoarding_NotPossible:
		return "NOT_ACCESSIBLE"
	default:
		return "UNKNOWN"
	}
}

// ParseFloatParam retrieves a float64 value from the provided URL query parameters.
// If the key is not present or the value is invalid, it returns 0 and updates the fieldErrors map.
// - params: URL query parameters.
// - key: The key to look for in the query parameters.
// - fieldErrors: A map to collect validation errors for fields.
// Returns:
// - The parsed float64 value (or 0 if invalid).
// - The updated fieldErrors map containing any validation errors.
func ParseFloatParam(params url.Values, key string, fieldErrors map[string][]string) (float64, map[string][]string) {
	if fieldErrors == nil {
		fieldErrors = make(map[string][]string)
	}

	val := params.Get(key)
	if val == "" {
		return 0, fieldErrors
	}

	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		fieldErrors[key] = append(fieldErrors[key], fmt.Sprintf("Invalid field value for field %q.", key))
	}
	return f, fieldErrors
}
