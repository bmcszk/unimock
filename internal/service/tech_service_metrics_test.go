package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/bmcszk/unimock/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTechService_ComprehensiveMetrics(t *testing.T) {
	// Create tech service
	techService := service.NewTechService(time.Now())
	ctx := context.Background()

	// Test data tracking for different endpoints and status codes
	testPath := "/api/users"
	statusCodes := []int{200, 404}

	// Simulate various requests
	for _, status := range statusCodes {
		for i := 0; i < 3; i++ {
			techService.IncrementRequestCount(ctx, testPath)
			techService.TrackResponse(ctx, testPath, status)
		}
	}

	// Get metrics
	metrics := techService.GetMetrics(ctx)

	// Verify overall request count
	requestCount, ok := metrics["request_count"].(int64)
	require.True(t, ok, "request_count should be int64")
	assert.Equal(t, int64(6), requestCount, "Expected 6 total requests")

	// Verify endpoint stats
	apiEndpoints, ok := metrics["api_endpoints"].(map[string]int64)
	require.True(t, ok, "api_endpoints should be map[string]int64")
	assert.Equal(t, int64(6), apiEndpoints[testPath], "Expected 6 requests for test path")

	// Verify status code stats
	statusCodeStats, ok := metrics["status_code_stats"].(map[string]map[string]int64)
	require.True(t, ok, "status_code_stats should be map[string]map[string]int64")
	
	pathStats := statusCodeStats[testPath]
	assert.Equal(t, int64(3), pathStats["200"], "Expected 3 responses with status 200")
	assert.Equal(t, int64(3), pathStats["404"], "Expected 3 responses with status 404")
}

func TestTechService_StatusCodeTrackingOnly(t *testing.T) {
	// Test tracking responses without incrementing request count
	techService := service.NewTechService(time.Now())
	ctx := context.Background()

	// Track responses without incrementing request count
	techService.TrackResponse(ctx, "/test/path", 200)
	techService.TrackResponse(ctx, "/test/path", 200)
	techService.TrackResponse(ctx, "/test/path", 404)

	metrics := techService.GetMetrics(ctx)

	// Request count should be 0 since we only tracked responses
	requestCount, ok := metrics["request_count"].(int64)
	require.True(t, ok, "request_count should be int64")
	assert.Equal(t, int64(0), requestCount)

	// But status code stats should be populated
	statusCodeStats, ok := metrics["status_code_stats"].(map[string]map[string]int64)
	require.True(t, ok, "status_code_stats should be map[string]map[string]int64")
	pathStats := statusCodeStats["/test/path"]
	assert.Equal(t, int64(2), pathStats["200"])
	assert.Equal(t, int64(1), pathStats["404"])
}

func TestTechService_ConcurrentAccess(t *testing.T) {
	// Test thread safety with concurrent access
	techService := service.NewTechService(time.Now())
	ctx := context.Background()

	// Number of goroutines
	numGoroutines := 10
	numRequestsPerGoroutine := 100

	// Channel to wait for all goroutines to complete
	done := make(chan bool, numGoroutines)

	// Start multiple goroutines
	for i := 0; i < numGoroutines; i++ {
		go func(_ int) {
			path := "/concurrent/test"
			for j := 0; j < numRequestsPerGoroutine; j++ {
				techService.IncrementRequestCount(ctx, path)
				techService.TrackResponse(ctx, path, 200)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify metrics
	metrics := techService.GetMetrics(ctx)
	requestCount, ok := metrics["request_count"].(int64)
	require.True(t, ok, "request_count should be int64")
	expectedTotal := int64(numGoroutines * numRequestsPerGoroutine)
	assert.Equal(t, expectedTotal, requestCount)

	apiEndpoints, ok := metrics["api_endpoints"].(map[string]int64)
	require.True(t, ok, "api_endpoints should be map[string]int64")
	assert.Equal(t, expectedTotal, apiEndpoints["/concurrent/test"])

	statusCodeStats, ok := metrics["status_code_stats"].(map[string]map[string]int64)
	require.True(t, ok, "status_code_stats should be map[string]map[string]int64")
	pathStats := statusCodeStats["/concurrent/test"]
	assert.Equal(t, expectedTotal, pathStats["200"])
}

func TestTechService_EmptyMetrics(t *testing.T) {
	// Test metrics when no requests have been tracked
	techService := service.NewTechService(time.Now())
	ctx := context.Background()

	metrics := techService.GetMetrics(ctx)

	requestCount, ok := metrics["request_count"].(int64)
	require.True(t, ok, "request_count should be int64")
	assert.Equal(t, int64(0), requestCount)

	apiEndpoints, ok := metrics["api_endpoints"].(map[string]int64)
	require.True(t, ok, "api_endpoints should be map[string]int64")
	assert.Empty(t, apiEndpoints)

	statusCodeStats, ok := metrics["status_code_stats"].(map[string]map[string]int64)
	require.True(t, ok, "status_code_stats should be map[string]map[string]int64")
	assert.Empty(t, statusCodeStats)
}