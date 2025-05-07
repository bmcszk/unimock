package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bmcszk/unimock/pkg/model"
)

func Example() {
	// Create a new client
	client, err := NewClient("http://localhost:8080")
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
	createdScenario, err := client.CreateScenario(ctx, scenario)
	if err != nil {
		log.Fatalf("Failed to create scenario: %v", err)
	}

	fmt.Printf("Created scenario with UUID: %s\n", createdScenario.UUID)

	// Get all scenarios
	scenarios, err := client.ListScenarios(ctx)
	if err != nil {
		log.Fatalf("Failed to list scenarios: %v", err)
	}

	fmt.Printf("Found %d scenarios\n", len(scenarios))

	// Get a specific scenario
	retrievedScenario, err := client.GetScenario(ctx, createdScenario.UUID)
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

	_, err = client.UpdateScenario(ctx, createdScenario.UUID, updatedScenario)
	if err != nil {
		log.Fatalf("Failed to update scenario: %v", err)
	}

	fmt.Println("Updated scenario successfully")

	// Delete the scenario
	err = client.DeleteScenario(ctx, createdScenario.UUID)
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

	_, err = client.GetScenario(ctx2, "some-uuid")
	if err != nil {
		fmt.Println("Request was canceled as expected")
	}
}
