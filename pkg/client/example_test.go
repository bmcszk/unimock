package client_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bmcszk/unimock/pkg/client"
	"github.com/bmcszk/unimock/pkg/model"
)

func Example() {
	// Create a new client
	clientInstance, err := client.NewClient("http://localhost:8080")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a new scenario
	scenario := &model.Scenario{
		RequestPath: "GET /api/users",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"users": [{"id": "123", "name": "Test User"}]}`,
	}

	// Create the scenario
	createdScenario, err := clientInstance.CreateScenario(ctx, *scenario)
	if err != nil {
		log.Fatalf("Failed to create scenario: %v", err)
	}

	fmt.Printf("Created scenario with UUID: %s\n", createdScenario.UUID)

	// Get all scenarios
	scenarios, err := clientInstance.ListScenarios(ctx)
	if err != nil {
		log.Fatalf("Failed to list scenarios: %v", err)
	}

	fmt.Printf("Found %d scenarios\n", len(scenarios))

	// Get a specific scenario
	retrievedScenario, err := clientInstance.GetScenario(ctx, createdScenario.UUID)
	if err != nil {
		log.Fatalf("Failed to get scenario: %v", err)
	}

	fmt.Printf("Retrieved scenario: %s %s\n", retrievedScenario.UUID, retrievedScenario.RequestPath)

	// Update the scenario
	updatedScenario := &model.Scenario{
		UUID:        createdScenario.UUID,
		RequestPath: "GET /api/users",
		StatusCode:  201,
		ContentType: "application/json",
		Data:        `{"users": [{"id": "123", "name": "Updated User"}]}`,
	}

	_, err = clientInstance.UpdateScenario(ctx, createdScenario.UUID, *updatedScenario)
	if err != nil {
		log.Fatalf("Failed to update scenario: %v", err)
	}

	fmt.Println("Updated scenario successfully")

	// Delete the scenario
	err = clientInstance.DeleteScenario(ctx, createdScenario.UUID)
	if err != nil {
		log.Fatalf("Failed to delete scenario: %v", err)
	}

	fmt.Println("Deleted scenario successfully")

	// Example using a context with cancellation
	ctx2, cancel2 := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel2() // Cancel after 100ms
	}()

	_, err = clientInstance.GetScenario(ctx2, "some-uuid")
	if err != nil {
		fmt.Println("Request was canceled as expected")
	}
}

// ExampleClient_universal_http_methods demonstrates using the client for universal HTTP requests
func ExampleClient_universal_http_methods() {
	// Create a client
	c, err := client.NewClient("http://localhost:8080")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Make a GET request
	resp, err := c.Get(ctx, "/api/users/123", map[string]string{
		"Authorization": "Bearer token",
	})
	if err != nil {
		log.Fatalf("GET request failed: %v", err)
	}
	fmt.Printf("GET Status: %d\n", resp.StatusCode)

	// Create a user with PostJSON
	userData := map[string]any{
		"id":   "123",
		"name": "John Doe",
	}
	resp, err = c.PostJSON(ctx, "/api/users", nil, userData)
	if err != nil {
		log.Fatalf("POST request failed: %v", err)
	}
	fmt.Printf("POST Status: %d\n", resp.StatusCode)

	// Check if resource exists with HEAD
	resp, err = c.Head(ctx, "/api/users/123", nil)
	if err != nil {
		log.Fatalf("HEAD request failed: %v", err)
	}
	fmt.Printf("HEAD Status: %d\n", resp.StatusCode)

	// Update with PutJSON
	updatedData := map[string]any{
		"id":   "123",
		"name": "Jane Doe",
	}
	resp, err = c.PutJSON(ctx, "/api/users/123", nil, updatedData)
	if err != nil {
		log.Fatalf("PUT request failed: %v", err)
	}
	fmt.Printf("PUT Status: %d\n", resp.StatusCode)

	// Delete the user
	resp, err = c.Delete(ctx, "/api/users/123", nil)
	if err != nil {
		log.Fatalf("DELETE request failed: %v", err)
	}
	fmt.Printf("DELETE Status: %d\n", resp.StatusCode)
}

func ExampleClient_HealthCheck() {
	// Create client
	c, err := client.NewClient("http://localhost:8080")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Check server health
	resp, err := c.HealthCheck(ctx)
	if err != nil {
		log.Fatalf("Health check failed: %v", err)
	}
	
	fmt.Printf("Health check status: %d\n", resp.StatusCode)
	if resp.StatusCode == 200 {
		fmt.Println("Server is healthy")
	}
}

// ExampleClient_mixed_usage demonstrates using both scenario management and HTTP requests
func ExampleClient_mixed_usage() {
	// Create a client
	c, err := client.NewClient("http://localhost:8080")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First, create a scenario for a specific endpoint
	scenario := &model.Scenario{
		RequestPath: "GET /api/special/endpoint",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"message": "This is a predefined response"}`,
	}

	created, err := c.CreateScenario(ctx, *scenario)
	if err != nil {
		log.Fatalf("Failed to create scenario: %v", err)
	}
	fmt.Printf("Created scenario: %s\n", created.UUID)

	// Now make a request to that endpoint and see the predefined response
	resp, err := c.Get(ctx, "/api/special/endpoint", nil)
	if err != nil {
		log.Fatalf("Failed to make request: %v", err)
	}
	fmt.Printf("Response: %s\n", string(resp.Body))

	// Test other endpoints with regular mock behavior
	userData := map[string]any{
		"id":   "user123",
		"name": "Regular User",
	}
	resp, err = c.PostJSON(ctx, "/api/users", nil, userData)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	fmt.Printf("User created with status: %d\n", resp.StatusCode)

	// Verify the user was created
	resp, err = c.Get(ctx, "/api/users/user123", nil)
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
	}
	fmt.Printf("User retrieved with status: %d\n", resp.StatusCode)

	// Clean up
	err = c.DeleteScenario(ctx, created.UUID)
	if err != nil {
		log.Fatalf("Failed to delete scenario: %v", err)
	}

	_, err = c.Delete(ctx, "/api/users/user123", nil)
	if err != nil {
		log.Fatalf("Failed to delete user: %v", err)
	}
	fmt.Println("Cleanup completed")
}
