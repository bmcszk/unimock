package model

// Scenario represents a predefined mock scenario for specific API requests
// Scenarios allow bypassing the normal mocking behavior for certain paths,
// enabling precise control over specific API responses.
type Scenario struct {
	// UUID is the unique identifier for the scenario
	// If not provided when creating, a UUID will be generated automatically
	UUID string `json:"uuid,omitempty"`

	// RequestPath defines which requests this scenario handles
	// Format: "METHOD /path" (e.g., "GET /api/users" or "POST /orders")
	// The path portion can contain wildcards (e.g., "GET /users/*")
	RequestPath string `json:"requestPath"`

	// StatusCode is the HTTP status code to return (e.g., 200, 201, 404, 500)
	StatusCode int `json:"statusCode"`

	// ContentType is the MIME type of the response (e.g., "application/json")
	ContentType string `json:"contentType"`

	// Location is the optional Location header value
	// Usually used with 201 Created responses
	Location string `json:"location,omitempty"`

	// Data is the response body to return
	// For JSON responses, this should be a valid JSON string
	Data string `json:"data"`

	// Headers is a map of HTTP headers to return with the scenario response
	Headers map[string]string `json:"headers,omitempty"`
}
