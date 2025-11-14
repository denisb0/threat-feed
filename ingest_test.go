package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockThreatWriter struct {
	hook func(IngestPayload) error
}

func (m *MockThreatWriter) Ingest(_ context.Context, payload IngestPayload) error {
	return m.hook(payload)
}

func setupIngestRouter(path string, h gin.HandlerFunc) *gin.Engine {
	// Set Gin to Release Mode to suppress verbose output during testing
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.POST(path, h)
	return r
}

func TestIngestHandlerSuccess(t *testing.T) {
	testPayload := IngestPayload{
		Feed:    "test-feed",
		Entries: []string{"1.2.3.4", "5.6.7.8"},
	}

	mockWriter := &MockThreatWriter{
		hook: func(payload IngestPayload) error {
			assert.Equal(t, testPayload.Feed, payload.Feed)
			assert.Equal(t, testPayload.Entries, payload.Entries)
			return nil
		},
	}

	handler := ingestHandler(mockWriter)
	router := setupIngestRouter("/ingest", handler)

	payloadBytes, err := json.Marshal(testPayload)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIngestHandlerBadRequest(t *testing.T) {
	mockWriter := &MockThreatWriter{
		hook: func(payload IngestPayload) error {
			t.Fatal("Ingest should not be called with invalid payload")
			return nil
		},
	}

	handler := ingestHandler(mockWriter)
	router := setupIngestRouter("/ingest", handler)

	// Send malformed JSON
	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestIngestHandlerThreatWriterError(t *testing.T) {
	testPayload := IngestPayload{
		Feed:    "test-feed",
		Entries: []string{"1.2.3.4", "5.6.7.8"},
	}

	mockWriter := &MockThreatWriter{
		hook: func(payload IngestPayload) error {
			return fmt.Errorf("storage error: database unavailable")
		},
	}

	handler := ingestHandler(mockWriter)
	router := setupIngestRouter("/ingest", handler)

	payloadBytes, err := json.Marshal(testPayload)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
