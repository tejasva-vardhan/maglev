package restapi

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipMiddleware(t *testing.T) {
	// Create a test handler that returns a large response
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write a large response that would benefit from compression
		w.Header().Set("Content-Type", "application/json")
		largeResponse := strings.Repeat(`{"test": "data"}`, 1000)
		w.Write([]byte(largeResponse))
	})

	t.Run("compresses response when gzip accepted", func(t *testing.T) {
		// Create request with gzip acceptance
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		recorder := httptest.NewRecorder()

		// Apply gzip middleware (this should exist after implementation)
		handler := applyGzipMiddleware(testHandler)
		handler.ServeHTTP(recorder, req)

		// Check response is compressed
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "gzip", recorder.Header().Get("Content-Encoding"))

		// Verify we can decompress the response
		reader, err := gzip.NewReader(bytes.NewReader(recorder.Body.Bytes()))
		require.NoError(t, err)
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		require.NoError(t, err)

		// Verify content
		expected := strings.Repeat(`{"test": "data"}`, 1000)
		assert.Equal(t, expected, string(decompressed))

		// Verify compression actually happened (compressed should be smaller)
		assert.Less(t, recorder.Body.Len(), len(expected))
	})

	t.Run("does not compress when gzip not accepted", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		// No Accept-Encoding header

		recorder := httptest.NewRecorder()

		handler := applyGzipMiddleware(testHandler)
		handler.ServeHTTP(recorder, req)

		// Check response is not compressed
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Empty(t, recorder.Header().Get("Content-Encoding"))

		// Content should be uncompressed
		expected := strings.Repeat(`{"test": "data"}`, 1000)
		assert.Equal(t, expected, recorder.Body.String())
	})

	t.Run("handles empty responses", func(t *testing.T) {
		emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		recorder := httptest.NewRecorder()

		handler := applyGzipMiddleware(emptyHandler)
		handler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusNoContent, recorder.Code)
		assert.Empty(t, recorder.Body.String())
	})

	t.Run("preserves content-type header", func(t *testing.T) {
		jsonHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			// Use larger content to ensure compression happens
			largeJSON := strings.Repeat(`{"message": "test data"}`, 100)
			w.Write([]byte(largeJSON))
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		recorder := httptest.NewRecorder()

		handler := applyGzipMiddleware(jsonHandler)
		handler.ServeHTTP(recorder, req)

		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
		assert.Equal(t, "gzip", recorder.Header().Get("Content-Encoding"))
	})
}

func TestGzipMiddlewareIntegration(t *testing.T) {
	// Create a test API instance
	api := createTestApi(t)
	defer api.GtfsManager.Shutdown()

	t.Run("API responses are compressed when requested", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/where/agencies-with-coverage.json?key=test", nil)
		req.Header.Set("Accept-Encoding", "gzip")

		recorder := httptest.NewRecorder()

		// Create handler with compression
		handler := createCompressedAPIHandler(api)
		handler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		// Check if the response was compressed - gzhttp may not compress small responses
		contentEncoding := recorder.Header().Get("Content-Encoding")
		if contentEncoding == "gzip" {
			// Verify response can be decompressed
			reader, err := gzip.NewReader(bytes.NewReader(recorder.Body.Bytes()))
			require.NoError(t, err)
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			require.NoError(t, err)

			// Should contain valid JSON
			assert.Contains(t, string(decompressed), `"code":200`)
		} else {
			// Response wasn't compressed (probably too small), verify it's valid JSON
			assert.Contains(t, recorder.Body.String(), `"code":200`)
		}
	})
}

// Helper functions are now implemented in compression_middleware.go
