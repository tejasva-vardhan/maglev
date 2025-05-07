package models

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestNewResponse(t *testing.T) {
	testCode := http.StatusCreated
	testData := map[string]string{"key": "value"}
	testText := "Resource Created"

	currentTimeBeforeCall := time.Now().UnixNano() / int64(time.Millisecond)
	response := NewResponse(testCode, testData, testText)
	currentTimeAfterCall := time.Now().UnixNano() / int64(time.Millisecond)

	assert.Equal(t, testCode, response.Code, "Response code should match input")
	assert.Equal(t, testData, response.Data, "Response data should match input")
	assert.Equal(t, testText, response.Text, "Response text should match input")
	assert.Equal(t, 2, response.Version, "Response version should be 2")
	assert.GreaterOrEqual(t, response.CurrentTime, currentTimeBeforeCall, "Response current time should be after or equal to time before call")
	assert.LessOrEqual(t, response.CurrentTime, currentTimeAfterCall, "Response current time should be before or equal to time after call")
	assert.InDelta(t, time.Now().UnixNano()/int64(time.Millisecond), response.CurrentTime, 100, "Response current time should be recent")
}

func TestNewEntryResponse(t *testing.T) {
	entryData := map[string]string{"id": "1", "name": "Test Entry"}
	references := NewEmptyReferences() // Assuming NewEmptyReferences is available

	response := NewEntryResponse(entryData, references)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "OK", response.Text)
	assert.Equal(t, 2, response.Version)
	assert.InDelta(t, time.Now().UnixNano()/int64(time.Millisecond), response.CurrentTime, 100)

	responseData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok, "Response data should be a map")
	assert.Equal(t, entryData, responseData["entry"], "Entry in response data should match input entry")
	assert.Equal(t, references, responseData["references"], "References in response data should match input references")
}

func TestNewOKResponse(t *testing.T) {
	testData := map[string]string{"status": "all good"}

	response := NewOKResponse(testData)

	assert.Equal(t, http.StatusOK, response.Code, "Response code should be StatusOK")
	assert.Equal(t, "OK", response.Text, "Response text should be 'OK'")
	assert.Equal(t, testData, response.Data, "Response data should match input")
	assert.Equal(t, 2, response.Version, "Response version should be 2")
	assert.InDelta(t, time.Now().UnixNano()/int64(time.Millisecond), response.CurrentTime, 100, "Response current time should be recent")
}

func TestNewListResponse(t *testing.T) {
	itemList := []string{"item1", "item2"}
	references := NewEmptyReferences()

	response := NewListResponse(itemList, references)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "OK", response.Text)
	assert.Equal(t, 2, response.Version)
	assert.InDelta(t, time.Now().UnixNano()/int64(time.Millisecond), response.CurrentTime, 100)

	responseData, ok := response.Data.(map[string]interface{})
	assert.True(t, ok, "Response data should be a map")
	assert.Equal(t, itemList, responseData["list"], "List in response data should match input list")
	assert.Equal(t, references, responseData["references"], "References in response data should match input references")
	assert.False(t, responseData["limitExceeded"].(bool), "limitExceeded should be false")
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
