/*
Package client provides a HTTP client for interacting with the Unimock API and mock endpoints.

The client allows programmatic access to:
1. Unimock server's scenario management functionality for setup, retrieve, update, and delete mock scenarios
2. Universal HTTP methods for making requests to any mock endpoint (GET, POST, PUT, DELETE, HEAD, PATCH, OPTIONS)

This makes it ideal for both test automation workflows and for testing applications against mocked services.

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

# Universal HTTP Methods

The client also provides universal HTTP methods for making requests to any mock endpoint:

Make a GET request:

	resp, err := client.Get(ctx, "/api/users/123", map[string]string{
		"Authorization": "Bearer token",
	})
	if err != nil {
		log.Fatalf("Failed to make GET request: %v", err)
	}
	
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Body: %s\n", string(resp.Body))

Make a POST request with JSON data:

	userData := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
	}
	
	resp, err := client.PostJSON(ctx, "/api/users", nil, userData)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

Make a PUT request with custom headers:

	headers := map[string]string{
		"Content-Type": "application/json",
		"X-API-Key":    "secret-key",
	}
	
	body := []byte(`{"name": "Updated Name"}`)
	resp, err := client.Put(ctx, "/api/users/123", headers, body)

Check if a resource exists with HEAD:

	resp, err := client.Head(ctx, "/api/users/123", nil)
	if err != nil {
		log.Fatalf("Failed to check resource: %v", err)
	}
	
	if resp.StatusCode == 200 {
		fmt.Println("Resource exists")
	}

Delete a resource:

	resp, err := client.Delete(ctx, "/api/users/123", nil)
	if err != nil {
		log.Fatalf("Failed to delete user: %v", err)
	}

All HTTP methods return a Response struct containing:

	type Response struct {
		StatusCode int         // HTTP status code (200, 404, etc.)
		Headers    http.Header // Response headers
		Body       []byte      // Response body content
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
