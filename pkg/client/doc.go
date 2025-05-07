/*
Package client provides a HTTP client for interacting with the Unimock API.

The client allows programmatic access to the Unimock server's scenario management
functionality, making it easier to setup, retrieve, update, and delete mock scenarios
as part of your test automation workflows.

# Basic Usage

Create a new client:

	client, err := client.NewClient("http://localhost:8080")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

All operations support context for cancellation and timeouts:

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

Create a new scenario:

	scenario := &model.Scenario{
		RequestPath: "GET /api/users",
		StatusCode:  200,
		ContentType: "application/json",
		Data:        `{"users": [{"id": "123", "name": "Test User"}]}`,
	}

	createdScenario, err := client.CreateScenario(ctx, scenario)
	if err != nil {
		log.Fatalf("Failed to create scenario: %v", err)
	}

List all scenarios:

	scenarios, err := client.ListScenarios(ctx)
	if err != nil {
		log.Fatalf("Failed to list scenarios: %v", err)
	}

Get a specific scenario:

	scenario, err := client.GetScenario(ctx, "scenario-uuid")
	if err != nil {
		log.Fatalf("Failed to get scenario: %v", err)
	}

Update a scenario:

	updatedScenario := &model.Scenario{
		UUID:        "scenario-uuid",
		RequestPath: "GET /api/users",
		StatusCode:  201,
		ContentType: "application/json",
		Data:        `{"users": [{"id": "123", "name": "Updated User"}]}`,
	}

	result, err := client.UpdateScenario(ctx, "scenario-uuid", updatedScenario)
	if err != nil {
		log.Fatalf("Failed to update scenario: %v", err)
	}

Delete a scenario:

	err = client.DeleteScenario(ctx, "scenario-uuid")
	if err != nil {
		log.Fatalf("Failed to delete scenario: %v", err)
	}

# Context Support

All client methods accept a context.Context parameter which can be used for:

1. Cancellation
2. Timeouts
3. Passing request-scoped values

Example with timeout:

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	scenario, err := client.GetScenario(ctx, "scenario-uuid")
	// The request will be cancelled if it takes longer than 5 seconds

Example with cancellation:

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the request after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_, err := client.GetScenario(ctx, "scenario-uuid")
	// The request will be cancelled after 100ms
*/
package client
