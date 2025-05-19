package utils

import (
	"fmt"
	"strings"
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
	return fmt.Sprintf("%s_%s", agencyID, codeID)
}
