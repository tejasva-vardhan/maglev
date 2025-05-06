package models

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestNewEmptyReferences(t *testing.T) {
	// Call the function to create an empty references model
	refs := NewEmptyReferences()

	// Check that all slices are initialized (not nil)
	if refs.Agencies == nil {
		t.Error("Agencies slice should be initialized, not nil")
	}
	if refs.Routes == nil {
		t.Error("Routes slice should be initialized, not nil")
	}
	if refs.Situations == nil {
		t.Error("Situations slice should be initialized, not nil")
	}
	if refs.StopTimes == nil {
		t.Error("StopTimes slice should be initialized, not nil")
	}
	if refs.Stops == nil {
		t.Error("Stops slice should be initialized, not nil")
	}
	if refs.Trips == nil {
		t.Error("Trips slice should be initialized, not nil")
	}

	// Check that all slices are empty
	if len(refs.Agencies) != 0 {
		t.Errorf("Expected Agencies to be empty, got length %d", len(refs.Agencies))
	}
	if len(refs.Routes) != 0 {
		t.Errorf("Expected Routes to be empty, got length %d", len(refs.Routes))
	}
	if len(refs.Situations) != 0 {
		t.Errorf("Expected Situations to be empty, got length %d", len(refs.Situations))
	}
	if len(refs.StopTimes) != 0 {
		t.Errorf("Expected StopTimes to be empty, got length %d", len(refs.StopTimes))
	}
	if len(refs.Stops) != 0 {
		t.Errorf("Expected Stops to be empty, got length %d", len(refs.Stops))
	}
	if len(refs.Trips) != 0 {
		t.Errorf("Expected Trips to be empty, got length %d", len(refs.Trips))
	}
}

func TestReferencesModelJSON(t *testing.T) {
	refs := NewEmptyReferences()
	refs.Agencies = append(refs.Agencies, AgencyReference{ID: "agency1"})
	refs.Routes = append(refs.Routes, map[string]string{"id": "route1"})

	// Marshal to JSON
	jsonData, err := json.Marshal(refs)
	if err != nil {
		t.Fatalf("Failed to marshal ReferencesModel to JSON: %v", err)
	}

	// Unmarshal back to a new struct
	var unmarshaledRefs ReferencesModel
	err = json.Unmarshal(jsonData, &unmarshaledRefs)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON to ReferencesModel: %v", err)
	}
	agency := unmarshaledRefs.Agencies[0]
	if agency.ID != "agency1" {
		t.Errorf("Expected agency id 'agency1', got %v", agency.ID)
	}
}

func TestNewResponse(t *testing.T) {
	// Create test data
	testCode := http.StatusOK
	testData := map[string]string{"test": "data"}
	testText := "Test Message"

	// Get current time before creating response for comparison
	beforeTime := time.Now().UnixNano() / int64(time.Millisecond)

	// Create the response
	response := NewResponse(testCode, testData, testText)

	// Get time after response creation
	afterTime := time.Now().UnixNano() / int64(time.Millisecond)

	// Test response fields
	if response.Code != testCode {
		t.Errorf("Expected response code %d, got %d", testCode, response.Code)
	}

	if response.Text != testText {
		t.Errorf("Expected response text %s, got %s", testText, response.Text)
	}

	if response.Version != 2 {
		t.Errorf("Expected response version 2, got %d", response.Version)
	}

	// Test that response time is reasonable
	if response.CurrentTime < beforeTime || response.CurrentTime > afterTime {
		t.Errorf("Response time %d is outside expected range [%d, %d]",
			response.CurrentTime, beforeTime, afterTime)
	}

	// Test that data was correctly set
	responseData, ok := response.Data.(map[string]string)
	if !ok {
		t.Error("Failed to cast response data to map[string]string")
	} else {
		if testValue, ok := responseData["test"]; !ok || testValue != "data" {
			t.Errorf("Expected response data {\"test\": \"data\"}, got %v", responseData)
		}
	}
}

func TestResponseModelJSON(t *testing.T) {
	// Create a response model with test data
	response := ResponseModel{
		Code:        http.StatusOK,
		CurrentTime: 1746324484528,
		Data:        map[string]string{"test": "data"},
		Text:        "Test Message",
		Version:     2,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal ResponseModel to JSON: %v", err)
	}

	// Unmarshal back to a new struct
	var unmarshaledResponse ResponseModel
	err = json.Unmarshal(jsonData, &unmarshaledResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON to ResponseModel: %v", err)
	}

	// Check field equality
	if unmarshaledResponse.Code != response.Code {
		t.Errorf("Expected code %d, got %d", response.Code, unmarshaledResponse.Code)
	}

	if unmarshaledResponse.CurrentTime != response.CurrentTime {
		t.Errorf("Expected currentTime %d, got %d",
			response.CurrentTime, unmarshaledResponse.CurrentTime)
	}

	if unmarshaledResponse.Text != response.Text {
		t.Errorf("Expected text %s, got %s", response.Text, unmarshaledResponse.Text)
	}

	if unmarshaledResponse.Version != response.Version {
		t.Errorf("Expected version %d, got %d", response.Version, unmarshaledResponse.Version)
	}

	// Check that data was correctly marshaled/unmarshaled
	responseData, ok := unmarshaledResponse.Data.(map[string]interface{})
	if !ok {
		t.Error("Failed to cast unmarshaled response data to map[string]interface{}")
	} else {
		if testValue, ok := responseData["test"].(string); !ok || testValue != "data" {
			t.Errorf("Expected response data {\"test\": \"data\"}, got %v", responseData)
		}
	}
}
