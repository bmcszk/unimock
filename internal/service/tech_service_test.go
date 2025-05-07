package service

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestTechService_GetHealthStatus(t *testing.T) {
	// Create service with fixed start time
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	service := NewTechService(startTime)

	// Get health status
	status := service.GetHealthStatus(context.Background())

	// Check status
	if status["status"] != "ok" {
		t.Errorf("status = %q, want %q", status["status"], "ok")
	}

	// Check uptime
	expectedUptime := time.Since(startTime).String()
	actualUptime := status["uptime"].(string)
	if !strings.HasPrefix(actualUptime, strings.Split(expectedUptime, ".")[0]) {
		t.Errorf("uptime = %q, want prefix %q", actualUptime, strings.Split(expectedUptime, ".")[0])
	}

	// Wait a bit and check again
	time.Sleep(100 * time.Millisecond)
	status = service.GetHealthStatus(context.Background())

	// Check status again
	if status["status"] != "ok" {
		t.Errorf("status = %q, want %q", status["status"], "ok")
	}

	// Check uptime again
	expectedUptime = time.Since(startTime).String()
	actualUptime = status["uptime"].(string)
	if !strings.HasPrefix(actualUptime, strings.Split(expectedUptime, ".")[0]) {
		t.Errorf("uptime = %q, want prefix %q", actualUptime, strings.Split(expectedUptime, ".")[0])
	}
}

func TestTechService_GetMetrics(t *testing.T) {
	// Create service with fixed start time
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	service := NewTechService(startTime)

	// Get initial metrics
	metrics := service.GetMetrics(context.Background())

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

	// Increment request counts
	service.IncrementRequestCount(context.Background(), "/api/users")
	service.IncrementRequestCount(context.Background(), "/api/users")
	service.IncrementRequestCount(context.Background(), "/api/users")
	service.IncrementRequestCount(context.Background(), "/api/orders")

	// Get updated metrics
	metrics = service.GetMetrics(context.Background())

	// Check total request count
	if metrics["request_count"] != int64(4) {
		t.Errorf("request_count = %d, want %d", metrics["request_count"], 4)
	}

	// Check endpoint stats
	endpoints, ok = metrics["api_endpoints"].(map[string]int64)
	if !ok {
		t.Error("api_endpoints is not a map[string]int64")
	}

	// Check specific endpoint counts
	expectedCounts := map[string]int64{
		"/api/users":  3,
		"/api/orders": 1,
	}

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

func TestTechService_IncrementRequestCount(t *testing.T) {
	// Create service with fixed start time
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	service := NewTechService(startTime)

	tests := []struct {
		name          string
		endpoint      string
		expectedCount int64
		expectedTotal int64
		expectedError bool
		errorContains string
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
			expectedTotal: 3,
		},
		{
			name:          "empty endpoint",
			endpoint:      "",
			expectedCount: 1,
			expectedTotal: 4,
		},
		{
			name:          "case-sensitive endpoint",
			endpoint:      "/API/users",
			expectedCount: 1,
			expectedTotal: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Increment request count
			service.IncrementRequestCount(context.Background(), tt.endpoint)

			// Get metrics
			metrics := service.GetMetrics(context.Background())

			// Check total request count
			if metrics["request_count"] != tt.expectedTotal {
				t.Errorf("total request_count = %d, want %d", metrics["request_count"], tt.expectedTotal)
			}

			// Check endpoint count
			endpoints, ok := metrics["api_endpoints"].(map[string]int64)
			if !ok {
				t.Error("api_endpoints is not a map[string]int64")
				return
			}

			// Check endpoint count
			if count, exists := endpoints[tt.endpoint]; !exists {
				t.Errorf("endpoint %q not found in metrics", tt.endpoint)
			} else if count != tt.expectedCount {
				t.Errorf("endpoint %q count = %d, want %d", tt.endpoint, count, tt.expectedCount)
			}
		})
	}
}
