package model

// Scenario represents a mock scenario resource
type Scenario struct {
	UUID        string `json:"uuid,omitempty"`
	RequestPath string `json:"requestPath"` // Format: "METHOD /path" (e.g., "GET /api/users")
	StatusCode  int    `json:"statusCode"`
	ContentType string `json:"contentType"`
	Location    string `json:"location,omitempty"`
	Data        string `json:"data"`
}
