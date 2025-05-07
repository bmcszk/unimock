package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bmcszk/unimock/pkg/model"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		expectError bool
	}{
		{
			name:        "valid URL",
			baseURL:     "http://localhost:8080",
			expectError: false,
		},
		{
			name:        "empty URL defaults to localhost",
			baseURL:     "",
			expectError: false,
		},
		{
			name:        "invalid URL",
			baseURL:     "://invalid-url",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClient(tc.baseURL)

			if tc.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Error("expected client, got nil")
				return
			}

			if tc.baseURL == "" && client.BaseURL.String() != DefaultBaseURL {
				t.Errorf("expected default base URL, got %s", client.BaseURL.String())
			} else if tc.baseURL != "" && client.BaseURL.String() != tc.baseURL {
				t.Errorf("expected base URL %s, got %s", tc.baseURL, client.BaseURL.String())
			}

			if client.HTTPClient == nil {
				t.Error("expected HTTP client, got nil")
			}
		})
	}
}

func TestClientOperations(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Test scenario data
		testScenario := &model.Scenario{
			UUID:        "test-uuid",
			RequestPath: "GET /api/test",
			StatusCode:  200,
			ContentType: "application/json",
			Data:        `{"message":"Test data"}`,
		}

		// Test scenario list
		testScenarios := []*model.Scenario{testScenario}

		// Set default content type
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/_uni/scenarios":
			// List scenarios
			json.NewEncoder(w).Encode(testScenarios)

		case r.Method == http.MethodGet && r.URL.Path == "/_uni/scenarios/test-uuid":
			// Get scenario
			json.NewEncoder(w).Encode(testScenario)

		case r.Method == http.MethodGet && r.URL.Path == "/_uni/scenarios/not-found":
			// Scenario not found
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not found"))

		case r.Method == http.MethodPost && r.URL.Path == "/_uni/scenarios":
			// Create scenario
			var scenario model.Scenario
			if err := json.NewDecoder(r.Body).Decode(&scenario); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Invalid request body"))
				return
			}
			// Set UUID if not provided
			if scenario.UUID == "" {
				scenario.UUID = "new-uuid"
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(scenario)

		case r.Method == http.MethodPut && r.URL.Path == "/_uni/scenarios/test-uuid":
			// Update scenario
			var scenario model.Scenario
			if err := json.NewDecoder(r.Body).Decode(&scenario); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Invalid request body"))
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(scenario)

		case r.Method == http.MethodPut && r.URL.Path == "/_uni/scenarios/not-found":
			// Update non-existent scenario
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not found"))

		case r.Method == http.MethodDelete && r.URL.Path == "/_uni/scenarios/test-uuid":
			// Delete scenario
			w.WriteHeader(http.StatusNoContent)

		case r.Method == http.MethodDelete && r.URL.Path == "/_uni/scenarios/not-found":
			// Delete non-existent scenario
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not found"))

		default:
			// Unknown endpoint
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Endpoint not found"))
		}
	}))
	defer server.Close()

	// Create a client with the test server URL
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("GetScenario", func(t *testing.T) {
		// Test getting an existing scenario
		scenario, err := client.GetScenario(ctx, "test-uuid")
		if err != nil {
			t.Errorf("Failed to get scenario: %v", err)
		}
		if scenario == nil {
			t.Fatal("Expected scenario, got nil")
		}
		if scenario.UUID != "test-uuid" {
			t.Errorf("Expected UUID test-uuid, got %s", scenario.UUID)
		}

		// Test getting a non-existent scenario
		_, err = client.GetScenario(ctx, "not-found")
		if err == nil {
			t.Error("Expected error for non-existent scenario, got nil")
		}
	})

	t.Run("ListScenarios", func(t *testing.T) {
		scenarios, err := client.ListScenarios(ctx)
		if err != nil {
			t.Errorf("Failed to list scenarios: %v", err)
		}
		if len(scenarios) != 1 {
			t.Errorf("Expected 1 scenario, got %d", len(scenarios))
		}
		if scenarios[0].UUID != "test-uuid" {
			t.Errorf("Expected UUID test-uuid, got %s", scenarios[0].UUID)
		}
	})

	t.Run("CreateScenario", func(t *testing.T) {
		newScenario := &model.Scenario{
			RequestPath: "POST /api/test",
			StatusCode:  201,
			ContentType: "application/json",
			Data:        `{"message":"New test data"}`,
		}

		created, err := client.CreateScenario(ctx, newScenario)
		if err != nil {
			t.Errorf("Failed to create scenario: %v", err)
		}
		if created == nil {
			t.Fatal("Expected created scenario, got nil")
		}
		if created.UUID != "new-uuid" {
			t.Errorf("Expected UUID new-uuid, got %s", created.UUID)
		}
		if created.RequestPath != newScenario.RequestPath {
			t.Errorf("Expected request path %s, got %s", newScenario.RequestPath, created.RequestPath)
		}
	})

	t.Run("UpdateScenario", func(t *testing.T) {
		updateScenario := &model.Scenario{
			UUID:        "test-uuid",
			RequestPath: "PUT /api/test",
			StatusCode:  200,
			ContentType: "application/json",
			Data:        `{"message":"Updated test data"}`,
		}

		// Test updating an existing scenario
		updated, err := client.UpdateScenario(ctx, "test-uuid", updateScenario)
		if err != nil {
			t.Errorf("Failed to update scenario: %v", err)
		}
		if updated == nil {
			t.Fatal("Expected updated scenario, got nil")
		}
		if updated.RequestPath != updateScenario.RequestPath {
			t.Errorf("Expected request path %s, got %s", updateScenario.RequestPath, updated.RequestPath)
		}

		// Test updating a non-existent scenario
		_, err = client.UpdateScenario(ctx, "not-found", updateScenario)
		if err == nil {
			t.Error("Expected error for non-existent scenario, got nil")
		}
	})

	t.Run("DeleteScenario", func(t *testing.T) {
		// Test deleting an existing scenario
		err := client.DeleteScenario(ctx, "test-uuid")
		if err != nil {
			t.Errorf("Failed to delete scenario: %v", err)
		}

		// Test deleting a non-existent scenario
		err = client.DeleteScenario(ctx, "not-found")
		if err == nil {
			t.Error("Expected error for non-existent scenario, got nil")
		}
	})
}

func TestClientContextCancellation(t *testing.T) {
	// Create a test server that delays for a bit
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep to simulate a delay
		time.Sleep(200 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	// Create a client with the test server URL
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create a context with immediate cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Try to make a request with canceled context
	_, err = client.GetScenario(ctx, "test-uuid")
	if err == nil {
		t.Error("Expected error due to context cancellation, got nil")
	}
}
