package model

// Scenario represents a mock scenario resource
type Scenario struct {
	UUID        string `json:"uuid,omitempty"`
	Path        string `json:"path"`
	StatusCode  int    `json:"statusCode"`
	ContentType string `json:"contentType"`
	Location    string `json:"location,omitempty"`
	Data        string `json:"data"`
}
