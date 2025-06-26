// Package model provides data structures for Unimock's HTTP mocking and scenarios.
package model

// UniData represents the data stored for a uni HTTP resource
type UniData struct {
	// Path of the resource (e.g., "/users/123")
	Path string `json:"path"`

	// IDs contains all identifiers associated with this resource
	// Multiple IDs allow a single resource to be accessible via different identifiers
	IDs []string `json:"ids,omitempty"`

	// Location header value for the resource
	// Usually used in POST responses to indicate where a new resource can be found
	Location string `json:"location,omitempty"`

	// ContentType specifies the MIME type of the data (e.g., "application/json")
	ContentType string `json:"content_type"`

	// Body contains the raw response data to be returned
	// For JSON responses, this is the serialized JSON body
	Body []byte `json:"body"`
}
