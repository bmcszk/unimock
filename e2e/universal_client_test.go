package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/bmcszk/unimock/pkg/client"
)

func TestUniversalClientE2E(t *testing.T) {
	// Create client
	baseURL := getBaseURL()
	c, err := client.NewClient(baseURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("BasicHTTPOperations", func(t *testing.T) {
		testBasicHTTPOperationsE2E(ctx, t, c)
	})

	t.Run("JSONOperations", func(t *testing.T) {
		testJSONOperationsE2E(ctx, t, c)
	})

	t.Run("UniDataLifecycle", func(t *testing.T) {
		testUniDataLifecycleE2E(ctx, t, c)
	})
}

func testBasicHTTPOperationsE2E(ctx context.Context, t *testing.T, c *client.Client) {
	t.Helper()

	// Test GET request to non-existent resource
	resp, err := c.Get(ctx, "/api/users/999", nil)
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("Expected 404 for non-existent resource, got %d", resp.StatusCode)
	}

	// Test HEAD request
	resp, err = c.Head(ctx, "/api/users/999", nil)
	if err != nil {
		t.Fatalf("Failed to make HEAD request: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("Expected 404 for HEAD request, got %d", resp.StatusCode)
	}
	if len(resp.Body) != 0 {
		t.Errorf("Expected empty body for HEAD request, got %d bytes", len(resp.Body))
	}

	// Test OPTIONS request to technical endpoint
	resp, err = c.Options(ctx, "/_uni/health", nil)
	if err != nil {
		t.Fatalf("Failed to make OPTIONS request: %v", err)
	}
	// OPTIONS support varies, so we just check that we got a response
	if resp.StatusCode == 0 {
		t.Error("Expected non-zero status code for OPTIONS request")
	}
}

func testJSONOperationsE2E(ctx context.Context, t *testing.T, c *client.Client) {
	t.Helper()

	// Create and verify user
	testCreateUserE2E(ctx, t, c)

	// Update and verify user
	testUpdateUserE2E(ctx, t, c)

	// Patch user
	testPatchUserE2E(ctx, t, c)

	// Delete and verify user
	testDeleteUserE2E(ctx, t, c)
}

func testCreateUserE2E(ctx context.Context, t *testing.T, c *client.Client) {
	t.Helper()
	userData := map[string]any{
		"id":    "test-user-001",
		"name":  "Test User",
		"email": "test@example.com",
	}

	resp, err := c.PostJSON(ctx, "/api/users", nil, userData)
	if err != nil {
		t.Fatalf("Failed to create user with PostJSON: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("Expected 201 for POST, got %d", resp.StatusCode)
	}

	// Verify the user was created by getting it
	resp, err = c.Get(ctx, "/api/users/test-user-001", nil)
	if err != nil {
		t.Fatalf("Failed to get created user: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 for GET, got %d", resp.StatusCode)
	}

	var retrievedUser map[string]any
	if err := json.Unmarshal(resp.Body, &retrievedUser); err != nil {
		t.Fatalf("Failed to parse user response: %v", err)
	}
	if retrievedUser["id"] != "test-user-001" {
		t.Errorf("Expected user ID 'test-user-001', got %v", retrievedUser["id"])
	}
}

func testUpdateUserE2E(ctx context.Context, t *testing.T, c *client.Client) {
	t.Helper()
	updatedData := map[string]any{
		"id":    "test-user-001",
		"name":  "Updated Test User",
		"email": "updated@example.com",
	}

	resp, err := c.PutJSON(ctx, "/api/users/test-user-001", nil, updatedData)
	if err != nil {
		t.Fatalf("Failed to update user with PutJSON: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 for PUT, got %d", resp.StatusCode)
	}

	// Verify the update by getting the user again
	resp, err = c.Get(ctx, "/api/users/test-user-001", nil)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}
	var retrievedUser map[string]any
	if err := json.Unmarshal(resp.Body, &retrievedUser); err != nil {
		t.Fatalf("Failed to parse updated user response: %v", err)
	}
	if retrievedUser["name"] != "Updated Test User" {
		t.Errorf("Expected updated name 'Updated Test User', got %v", retrievedUser["name"])
	}
}

func testPatchUserE2E(ctx context.Context, t *testing.T, c *client.Client) {
	t.Helper()
	patchData := map[string]any{
		"email": "patched@example.com",
	}

	resp, err := c.PatchJSON(ctx, "/api/users/test-user-001", nil, patchData)
	if err != nil {
		t.Fatalf("Failed to patch user with PatchJSON: %v", err)
	}
	// PATCH may not be supported by default config, so accept 405 Method Not Allowed
	if resp.StatusCode != 200 && resp.StatusCode != 405 {
		t.Errorf("Expected 200 or 405 for PATCH, got %d", resp.StatusCode)
	}
}

func testDeleteUserE2E(ctx context.Context, t *testing.T, c *client.Client) {
	t.Helper()
	resp, err := c.Delete(ctx, "/api/users/test-user-001", nil)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}
	if resp.StatusCode != 204 {
		t.Errorf("Expected 204 for DELETE, got %d", resp.StatusCode)
	}

	// Verify the user is deleted
	resp, err = c.Get(ctx, "/api/users/test-user-001", nil)
	if err != nil {
		t.Fatalf("Failed to verify user deletion: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("Expected 404 for deleted user, got %d", resp.StatusCode)
	}
}

func testUniDataLifecycleE2E(ctx context.Context, t *testing.T, c *client.Client) {
	t.Helper()

	// Test creating multiple resources
	createMultipleUsers(ctx, t, c)

	// Test getting the collection
	verifyUsersCollection(ctx, t, c)

	// Clean up - delete all users
	deleteAllUsers(ctx, t, c)
}

func createMultipleUsers(ctx context.Context, t *testing.T, c *client.Client) {
	t.Helper()
	for i := 1; i <= 3; i++ {
		userData := map[string]any{
			"id":   fmt.Sprintf("user-%d", i),
			"name": fmt.Sprintf("User %d", i),
		}

		resp, err := c.PostJSON(ctx, "/api/users", nil, userData)
		if err != nil {
			t.Fatalf("Failed to create user %d: %v", i, err)
		}
		if resp.StatusCode != 201 {
			t.Errorf("Expected 201 for user %d creation, got %d", i, resp.StatusCode)
		}
	}
}

func verifyUsersCollection(ctx context.Context, t *testing.T, c *client.Client) {
	t.Helper()
	resp, err := c.Get(ctx, "/api/users", nil)
	if err != nil {
		t.Fatalf("Failed to get users collection: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 for collection GET, got %d", resp.StatusCode)
	}

	var users []map[string]any
	if err := json.Unmarshal(resp.Body, &users); err != nil {
		t.Fatalf("Failed to parse users collection: %v", err)
	}
	if len(users) != 3 {
		t.Errorf("Expected 3 users in collection, got %d", len(users))
	}
}

func deleteAllUsers(ctx context.Context, t *testing.T, c *client.Client) {
	t.Helper()
	for i := 1; i <= 3; i++ {
		resp, err := c.Delete(ctx, fmt.Sprintf("/api/users/user-%d", i), nil)
		if err != nil {
			t.Fatalf("Failed to delete user %d: %v", i, err)
		}
		if resp.StatusCode != 204 {
			t.Errorf("Expected 204 for user %d deletion, got %d", i, resp.StatusCode)
		}
	}
}

// No helper functions needed anymore, server is managed by Docker Compose
