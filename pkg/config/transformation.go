package config

import (
	"github.com/bmcszk/unimock/pkg/model"
)

// RequestTransformFunc defines a function type for transforming request data before processing.
// It receives the UniData and should return the transformed UniData or an error if transformation fails.
// Returning an error will result in a 500 Internal Server Error being sent to the client.
type RequestTransformFunc func(data model.UniData) (model.UniData, error)

// ResponseTransformFunc defines a function type for transforming response data after processing.
// It receives the UniData to be returned and should return the transformed UniData or an error.
// Returning an error will result in a 500 Internal Server Error being sent to the client.
type ResponseTransformFunc func(data model.UniData) (model.UniData, error)

// TransformationConfig holds the transformation functions and configuration for a section.
// This configuration is only available when using Unimock as a library and cannot be
// configured via YAML files.
type TransformationConfig struct {
	// RequestTransforms contains a list of request transformation functions to be executed in order.
	// If any transformation function returns an error, a 500 Internal Server Error is returned.
	RequestTransforms []RequestTransformFunc

	// ResponseTransforms contains a list of response transformation functions to be executed in order.
	// If any transformation function returns an error, a 500 Internal Server Error is returned.
	ResponseTransforms []ResponseTransformFunc
}

// NewTransformationConfig creates a new TransformationConfig with default settings
func NewTransformationConfig() *TransformationConfig {
	return &TransformationConfig{
		RequestTransforms:  make([]RequestTransformFunc, 0),
		ResponseTransforms: make([]ResponseTransformFunc, 0),
	}
}

// AddRequestTransform adds a request transformation function to the configuration
func (tc *TransformationConfig) AddRequestTransform(transform RequestTransformFunc) {
	if tc.RequestTransforms == nil {
		tc.RequestTransforms = make([]RequestTransformFunc, 0)
	}
	tc.RequestTransforms = append(tc.RequestTransforms, transform)
}

// AddResponseTransform adds a response transformation function to the configuration
func (tc *TransformationConfig) AddResponseTransform(transform ResponseTransformFunc) {
	if tc.ResponseTransforms == nil {
		tc.ResponseTransforms = make([]ResponseTransformFunc, 0)
	}
	tc.ResponseTransforms = append(tc.ResponseTransforms, transform)
}

// HasRequestTransforms returns true if any request transformations are configured
func (tc *TransformationConfig) HasRequestTransforms() bool {
	return tc != nil && len(tc.RequestTransforms) > 0
}

// HasResponseTransforms returns true if any response transformations are configured
func (tc *TransformationConfig) HasResponseTransforms() bool {
	return tc != nil && len(tc.ResponseTransforms) > 0
}

// HasAnyTransforms returns true if any transformations are configured
func (tc *TransformationConfig) HasAnyTransforms() bool {
	return tc.HasRequestTransforms() || tc.HasResponseTransforms()
}