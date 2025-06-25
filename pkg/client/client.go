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

	// HTTP client timeout
	httpClientTimeout = 10 * time.Second

	// HTTP status code boundaries
	httpStatusOKMin = 200
	httpStatusOKMax = 300

	// Common error message templates
	msgFailedCreateRequest = "failed to create request: %w"
	msgFailedSendRequest   = "failed to send request: %w"
	msgServerError         = "server returned error status %d: %s"
	msgFailedParseResponse = "failed to parse response: %w"
)

// Client is an HTTP client for interacting with the Unimock API
type Client struct {
	// BaseURL is the base URL of the Unimock server
	BaseURL *url.URL

	// HTTPClient is the underlying HTTP client used to make requests
	HTTPClient *http.Client
}

// Response represents an HTTP response from the server
type Response struct {
	// StatusCode is the HTTP status code (200, 404, etc.)
	StatusCode int

	// Headers contains the response headers
	Headers http.Header

	// Body contains the response body content
	Body []byte
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
			Timeout: httpClientTimeout,
		},
	}, nil
}

// ========================================
// Universal HTTP Methods
// ========================================

// Get performs a GET request to the specified path
func (c *Client) Get(ctx context.Context, urlPath string, headers map[string]string) (*Response, error) {
	return c.doRequest(ctx, http.MethodGet, urlPath, headers, nil)
}

// Head performs a HEAD request to the specified path
func (c *Client) Head(ctx context.Context, urlPath string, headers map[string]string) (*Response, error) {
	return c.doRequest(ctx, http.MethodHead, urlPath, headers, nil)
}

// Post performs a POST request to the specified path with the given body
func (c *Client) Post(ctx context.Context, urlPath string, headers map[string]string, body []byte) (*Response, error) {
	return c.doRequest(ctx, http.MethodPost, urlPath, headers, body)
}

// Put performs a PUT request to the specified path with the given body
func (c *Client) Put(ctx context.Context, urlPath string, headers map[string]string, body []byte) (*Response, error) {
	return c.doRequest(ctx, http.MethodPut, urlPath, headers, body)
}

// Delete performs a DELETE request to the specified path
func (c *Client) Delete(ctx context.Context, urlPath string, headers map[string]string) (*Response, error) {
	return c.doRequest(ctx, http.MethodDelete, urlPath, headers, nil)
}

// Patch performs a PATCH request to the specified path with the given body
func (c *Client) Patch(ctx context.Context, urlPath string, headers map[string]string, body []byte) (*Response, error) {
	return c.doRequest(ctx, http.MethodPatch, urlPath, headers, body)
}

// Options performs an OPTIONS request to the specified path
func (c *Client) Options(ctx context.Context, urlPath string, headers map[string]string) (*Response, error) {
	return c.doRequest(ctx, http.MethodOptions, urlPath, headers, nil)
}

// PostJSON performs a POST request with JSON content
func (c *Client) PostJSON(ctx context.Context, urlPath string, headers map[string]string, data any) (*Response, error) {
	return c.doJSONRequest(ctx, http.MethodPost, urlPath, headers, data)
}

// PutJSON performs a PUT request with JSON content
func (c *Client) PutJSON(ctx context.Context, urlPath string, headers map[string]string, data any) (*Response, error) {
	return c.doJSONRequest(ctx, http.MethodPut, urlPath, headers, data)
}

// PatchJSON performs a PATCH request with JSON content
func (c *Client) PatchJSON(
	ctx context.Context,
	urlPath string,
	headers map[string]string,
	data any,
) (*Response, error) {
	return c.doJSONRequest(ctx, http.MethodPatch, urlPath, headers, data)
}

// doRequest is the internal method that performs HTTP requests
func (c *Client) doRequest(
	ctx context.Context,
	method, urlPath string,
	headers map[string]string,
	body []byte,
) (*Response, error) {
	// Build the request URL
	requestURL := c.buildRequestURL(urlPath)

	// Create request body reader
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, method, requestURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf(msgFailedCreateRequest, err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(msgFailedSendRequest, err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Create response object
	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
	}, nil
}

// doJSONRequest performs HTTP requests with JSON payloads
func (c *Client) doJSONRequest(
	ctx context.Context,
	method, urlPath string,
	headers map[string]string,
	data any,
) (*Response, error) {
	// Serialize data to JSON
	jsonBody, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize data to JSON: %w", err)
	}

	// Create headers map if nil
	if headers == nil {
		headers = make(map[string]string)
	}

	// Set Content-Type header
	headers["Content-Type"] = "application/json"

	// Perform the request
	return c.doRequest(ctx, method, urlPath, headers, jsonBody)
}

// HealthCheck performs a health check on the Unimock server
func (c *Client) HealthCheck(ctx context.Context) (*Response, error) {
	return c.Get(ctx, "/_uni/health", nil)
}

// buildRequestURL builds the complete URL for a request
func (c *Client) buildRequestURL(requestPath string) string {
	// If path is an absolute URL, parse it and use it directly
	if parsedPath, err := url.Parse(requestPath); err == nil && parsedPath.IsAbs() {
		return requestPath
	}

	// Build URL relative to base URL
	u := *c.BaseURL
	u.Path = requestPath
	return u.String()
}

// ========================================
// Scenario Management
// ========================================

// CreateScenario creates a new scenario
func (c *Client) CreateScenario(ctx context.Context, scenario *model.Scenario) (*model.Scenario, error) {
	requestURL := c.buildURL(scenarioBasePath)

	// Serialize the scenario to JSON
	body, err := json.Marshal(scenario)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize scenario: %w", err)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf(msgFailedCreateRequest, err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(msgFailedSendRequest, err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Handle error responses
	if resp.StatusCode < httpStatusOKMin || resp.StatusCode >= httpStatusOKMax {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(msgServerError, resp.StatusCode, string(respBody))
	}

	// Parse the response
	var createdScenario model.Scenario
	if err := json.NewDecoder(resp.Body).Decode(&createdScenario); err != nil {
		return nil, fmt.Errorf(msgFailedParseResponse, err)
	}

	return &createdScenario, nil
}

// GetScenario gets a scenario by UUID
func (c *Client) GetScenario(ctx context.Context, uuid string) (*model.Scenario, error) {
	requestURL := c.buildURL(path.Join(scenarioBasePath, uuid))

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf(msgFailedCreateRequest, err)
	}

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(msgFailedSendRequest, err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Handle error responses
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("scenario not found: %s", uuid)
	}
	if resp.StatusCode < httpStatusOKMin || resp.StatusCode >= httpStatusOKMax {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(msgServerError, resp.StatusCode, string(respBody))
	}

	// Parse the response
	var scenario model.Scenario
	if err := json.NewDecoder(resp.Body).Decode(&scenario); err != nil {
		return nil, fmt.Errorf(msgFailedParseResponse, err)
	}

	return &scenario, nil
}

// ListScenarios gets all scenarios
func (c *Client) ListScenarios(ctx context.Context) ([]*model.Scenario, error) {
	requestURL := c.buildURL(scenarioBasePath)

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf(msgFailedCreateRequest, err)
	}

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(msgFailedSendRequest, err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Handle error responses
	if resp.StatusCode < httpStatusOKMin || resp.StatusCode >= httpStatusOKMax {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(msgServerError, resp.StatusCode, string(respBody))
	}

	// Parse the response
	var scenarios []*model.Scenario
	if err := json.NewDecoder(resp.Body).Decode(&scenarios); err != nil {
		return nil, fmt.Errorf(msgFailedParseResponse, err)
	}

	return scenarios, nil
}

// UpdateScenario updates an existing scenario
func (c *Client) UpdateScenario(ctx context.Context, uuid string, scenario *model.Scenario) (*model.Scenario, error) {
	requestURL := c.buildURL(path.Join(scenarioBasePath, uuid))

	// Serialize the scenario to JSON
	body, err := json.Marshal(scenario)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize scenario: %w", err)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, requestURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf(msgFailedCreateRequest, err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(msgFailedSendRequest, err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Handle error responses
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("scenario not found: %s", uuid)
	}
	if resp.StatusCode < httpStatusOKMin || resp.StatusCode >= httpStatusOKMax {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(msgServerError, resp.StatusCode, string(respBody))
	}

	// Parse the response
	var updatedScenario model.Scenario
	if err := json.NewDecoder(resp.Body).Decode(&updatedScenario); err != nil {
		return nil, fmt.Errorf(msgFailedParseResponse, err)
	}

	return &updatedScenario, nil
}

// DeleteScenario deletes a scenario by UUID
func (c *Client) DeleteScenario(ctx context.Context, uuid string) error {
	requestURL := c.buildURL(path.Join(scenarioBasePath, uuid))

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, requestURL, nil)
	if err != nil {
		return fmt.Errorf(msgFailedCreateRequest, err)
	}

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf(msgFailedSendRequest, err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Handle error responses
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("scenario not found: %s", uuid)
	}
	if resp.StatusCode < httpStatusOKMin || resp.StatusCode >= httpStatusOKMax {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf(msgServerError, resp.StatusCode, string(respBody))
	}

	return nil
}

// Helper method to build a URL
func (c *Client) buildURL(urlPath string) string {
	u := *c.BaseURL
	u.Path = urlPath
	return u.String()
}
