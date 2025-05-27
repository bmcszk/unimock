package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/bmcszk/unimock/pkg/model"
)

const (
	// DefaultBaseURL is the default base URL for the Unimock server
	DefaultBaseURL = "http://localhost:8080"

	// scenarioBasePath is the base path for the scenario API
	scenarioBasePath = "/_uni/scenarios"
)

// Client is an HTTP client for interacting with the Unimock API
type Client struct {
	// BaseURL is the base URL of the Unimock server
	BaseURL *url.URL

	// HTTPClient is the underlying HTTP client used to make requests
	HTTPClient *http.Client
}

// NewClient creates a new client with the given base URL
func NewClient(baseURL string) (*Client, error) {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	return &Client{
		BaseURL: parsedURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// CreateScenario creates a new scenario
func (c *Client) CreateScenario(ctx context.Context, scenario *model.Scenario) (*model.Scenario, error) {
	url := c.buildURL(scenarioBasePath)

	// Serialize the scenario to JSON
	body, err := json.Marshal(scenario)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize scenario: %w", err)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Handle error responses
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned error status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse the response
	var createdScenario model.Scenario
	if err := json.NewDecoder(resp.Body).Decode(&createdScenario); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &createdScenario, nil
}

// GetScenario gets a scenario by UUID
func (c *Client) GetScenario(ctx context.Context, uuid string) (*model.Scenario, error) {
	url := c.buildURL(path.Join(scenarioBasePath, uuid))

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Handle error responses
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("scenario not found: %s", uuid)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned error status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse the response
	var scenario model.Scenario
	if err := json.NewDecoder(resp.Body).Decode(&scenario); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &scenario, nil
}

// ListScenarios gets all scenarios
func (c *Client) ListScenarios(ctx context.Context) ([]*model.Scenario, error) {
	url := c.buildURL(scenarioBasePath)

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Handle error responses
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned error status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse the response
	var scenarios []*model.Scenario
	if err := json.NewDecoder(resp.Body).Decode(&scenarios); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return scenarios, nil
}

// UpdateScenario updates an existing scenario
func (c *Client) UpdateScenario(ctx context.Context, uuid string, scenario *model.Scenario) (*model.Scenario, error) {
	url := c.buildURL(path.Join(scenarioBasePath, uuid))

	// Serialize the scenario to JSON
	body, err := json.Marshal(scenario)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize scenario: %w", err)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Handle error responses
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("scenario not found: %s", uuid)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned error status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse the response
	var updatedScenario model.Scenario
	if err := json.NewDecoder(resp.Body).Decode(&updatedScenario); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &updatedScenario, nil
}

// DeleteScenario deletes a scenario by UUID
func (c *Client) DeleteScenario(ctx context.Context, uuid string) error {
	url := c.buildURL(path.Join(scenarioBasePath, uuid))

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Handle error responses
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("scenario not found: %s", uuid)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// Helper method to build a URL
func (c *Client) buildURL(path string) string {
	u := *c.BaseURL
	u.Path = path
	return u.String()
}
