package service_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/bmcszk/unimock/internal/service"
)

func TestTechService_GetHealthStatus(t *testing.T) {
	// Create service with fixed start time
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	techSvc := service.NewTechService(startTime)

	// Get health status
	status := techSvc.GetHealthStatus(context.Background())

	// Check status
	if status["status"] != "ok" {
		t.Errorf("status = %q, want %q", status["status"], "ok")
	}

	// Check uptime
	expectedUptime := time.Since(startTime).String()
	actualUptime, ok := status["uptime"].(string)
	if !ok {
		t.Error("uptime is not a string")
		return
	}
	if !strings.HasPrefix(actualUptime, strings.Split(expectedUptime, ".")[0]) {
		t.Errorf("uptime = %q, want prefix %q", actualUptime, strings.Split(expectedUptime, ".")[0])
	}

	// Wait a bit and check again
	time.Sleep(100 * time.Millisecond) //nolint:gomnd
	status = techSvc.GetHealthStatus(context.Background())

	// Check status again
	if status["status"] != "ok" {
		t.Errorf("status = %q, want %q", status["status"], "ok")
	}

	// Check uptime again
	expectedUptime = time.Since(startTime).String()
	actualUptime, ok = status["uptime"].(string)
	if !ok {
		t.Error("uptime is not a string")
		return
	}
	if !strings.HasPrefix(actualUptime, strings.Split(expectedUptime, ".")[0]) {
		t.Errorf("uptime = %q, want prefix %q", actualUptime, strings.Split(expectedUptime, ".")[0])
	}
}

// Helper function to validate initial metrics
func validateInitialMetrics(t *testing.T, metrics map[string]any) {
	t.Helper()
	// Check initial request count
	if metrics["request_count"] != int64(0) {
		t.Errorf("request_count = %d, want %d", metrics["request_count"], 0)
	}

	// Check initial endpoint stats
	endpoints, ok := metrics["api_endpoints"].(map[string]int64)
	if !ok {
		t.Error("api_endpoints is not a map[string]int64")
	}
	if len(endpoints) != 0 {
		t.Errorf("api_endpoints has %d entries, want 0", len(endpoints))
	}
}

// Helper function to validate endpoint counts
func validateEndpointCounts(t *testing.T, endpoints map[string]int64, expectedCounts map[string]int64) {
	t.Helper()
	for endpoint, expectedCount := range expectedCounts {
		if count, exists := endpoints[endpoint]; !exists {
			t.Errorf("endpoint %q not found in metrics", endpoint)
		} else if count != expectedCount {
			t.Errorf("endpoint %q count = %d, want %d", endpoint, count, expectedCount)
		}
	}

	// Check for unexpected endpoints
	for endpoint := range endpoints {
		if _, exists := expectedCounts[endpoint]; !exists {
			t.Errorf("unexpected endpoint %q in metrics", endpoint)
		}
	}
}

func TestTechService_GetMetrics(t *testing.T) {
	// Create service with fixed start time
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	techSvc := service.NewTechService(startTime)

	// Get initial metrics
	metrics := techSvc.GetMetrics(context.Background())
	validateInitialMetrics(t, metrics)

	// Increment request counts
	techSvc.IncrementRequestCount(context.Background(), "/api/users")
	techSvc.IncrementRequestCount(context.Background(), "/api/users")
	techSvc.IncrementRequestCount(context.Background(), "/api/users")
	techSvc.IncrementRequestCount(context.Background(), "/api/orders")

	// Get updated metrics
	metrics = techSvc.GetMetrics(context.Background())

	// Check total request count
	if metrics["request_count"] != int64(4) { //nolint:gomnd
		t.Errorf("request_count = %d, want %d", metrics["request_count"], 4)
	}

	// Check endpoint stats
	endpoints, ok := metrics["api_endpoints"].(map[string]int64)
	if !ok {
		t.Error("api_endpoints is not a map[string]int64")
		return
	}

	// Check specific endpoint counts
	expectedCounts := map[string]int64{
		"/api/users":  3, //nolint:gomnd
		"/api/orders": 1,
	}

	validateEndpointCounts(t, endpoints, expectedCounts)
}

// Helper function to get test cases for IncrementRequestCount
func getIncrementRequestCountTestCases() []struct {
	name          string
	endpoint      string
	expectedCount int64
	expectedTotal int64
} {
	return []struct {
		name          string
		endpoint      string
		expectedCount int64
		expectedTotal int64
	}{
		{
			name:          "valid endpoint",
			endpoint:      "/api/users",
			expectedCount: 1,
			expectedTotal: 1,
		},
		{
			name:          "same endpoint again",
			endpoint:      "/api/users",
			expectedCount: 2,
			expectedTotal: 2,
		},
		{
			name:          "different endpoint",
			endpoint:      "/api/orders",
			expectedCount: 1,
			expectedTotal: 3, //nolint:mnd
		},
		{
			name:          "empty endpoint",
			endpoint:      "",
			expectedCount: 1,
			expectedTotal: 4, //nolint:mnd
		},
		{
			name:          "case-sensitive endpoint",
			endpoint:      "/API/users",
			expectedCount: 1,
			expectedTotal: 5, //nolint:mnd
		},
	}
}

// Helper function to validate request count metrics
func validateRequestCountMetrics(
	t *testing.T, metrics map[string]any, endpoint string, expectedCount, expectedTotal int64,
) {
	t.Helper()
	// Check total request count
	if metrics["request_count"] != expectedTotal {
		t.Errorf("total request_count = %d, want %d", metrics["request_count"], expectedTotal)
	}

	// Check endpoint count
	endpoints, ok := metrics["api_endpoints"].(map[string]int64)
	if !ok {
		t.Error("api_endpoints is not a map[string]int64")
		return
	}

	// Check endpoint count
	if count, exists := endpoints[endpoint]; !exists {
		t.Errorf("endpoint %q not found in metrics", endpoint)
	} else if count != expectedCount {
		t.Errorf("endpoint %q count = %d, want %d", endpoint, count, expectedCount)
	}
}

func TestTechService_IncrementRequestCount(t *testing.T) {
	// Create service with fixed start time
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	techSvc := service.NewTechService(startTime)

	tests := getIncrementRequestCountTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Increment request count
			techSvc.IncrementRequestCount(context.Background(), tt.endpoint)

			// Get metrics
			metrics := techSvc.GetMetrics(context.Background())

			validateRequestCountMetrics(t, metrics, tt.endpoint, tt.expectedCount, tt.expectedTotal)
		})
	}
}
