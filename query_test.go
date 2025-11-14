package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockThreatReader struct {
	hook func(string) (QueryResponse, error)
}

func (m *MockThreatReader) Query(_ context.Context, q string) (QueryResponse, error) {
	return m.hook(q)
}

func setupRouter(path string, h gin.HandlerFunc) *gin.Engine {
	// Set Gin to Release Mode to suppress verbose output during testing
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.GET(path, h)
	return r
}

func TestQueryThreat(t *testing.T) {
	t.Parallel()
	testIP := "1.2.3.4"
	mockReader := MockThreatReader{
		hook: func(q string) (QueryResponse, error) {
			assert.Equal(t, testIP, q)
			return QueryResponse{
				Threat: true,
				Feed:   "test-feed",
			}, nil
		},
	}

	handler := queryHandler(&mockReader)
	router := setupRouter("/query", handler)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/query?ip=%s", testIP), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response QueryResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, QueryResponse{
		Threat: true,
		Feed:   "test-feed",
	}, response)
}

func TestQueryNoThreat(t *testing.T) {
	t.Parallel()
	testIP := "1.2.3.4"
	mockReader := MockThreatReader{
		hook: func(q string) (QueryResponse, error) {
			assert.Equal(t, testIP, q)
			return QueryResponse{}, nil
		},
	}

	handler := queryHandler(&mockReader)
	router := setupRouter("/query", handler)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/query?ip=%s", testIP), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response QueryResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, QueryResponse{}, response)
}

func TestQueryStorageError(t *testing.T) {
	t.Parallel()
	testIP := "1.2.3.4"
	mockReader := MockThreatReader{
		hook: func(q string) (QueryResponse, error) {
			assert.Equal(t, testIP, q)
			return QueryResponse{}, fmt.Errorf("storage error: database unavailable")
		},
	}

	handler := queryHandler(&mockReader)
	router := setupRouter("/query", handler)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/query?ip=%s", testIP), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
