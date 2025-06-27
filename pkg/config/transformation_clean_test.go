package config_test

import (
	"testing"

	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestNewTransformationConfig(t *testing.T) {
	cfg := config.NewTransformationConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, 0, len(cfg.RequestTransforms))
	assert.Equal(t, 0, len(cfg.ResponseTransforms))
}

func TestTransformationConfig_AddRequestTransform(t *testing.T) {
	cfg := config.NewTransformationConfig()

	// Test adding to initialized config
	transform1 := func(data model.UniData) (model.UniData, error) {
		return data, nil
	}

	cfg.AddRequestTransform(transform1)
	assert.Equal(t, 1, len(cfg.RequestTransforms))

	// Test adding to nil slice
	cfg.RequestTransforms = nil
	cfg.AddRequestTransform(transform1)
	assert.Equal(t, 1, len(cfg.RequestTransforms))
}

func TestTransformationConfig_AddResponseTransform(t *testing.T) {
	cfg := config.NewTransformationConfig()

	// Test adding to initialized config
	transform1 := func(data model.UniData) (model.UniData, error) {
		return data, nil
	}

	cfg.AddResponseTransform(transform1)
	assert.Equal(t, 1, len(cfg.ResponseTransforms))

	// Test adding to nil slice
	cfg.ResponseTransforms = nil
	cfg.AddResponseTransform(transform1)
	assert.Equal(t, 1, len(cfg.ResponseTransforms))
}

func TestTransformationConfig_HasRequestTransforms(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.TransformationConfig
		expected bool
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: false,
		},
		{
			name:     "empty config",
			config:   config.NewTransformationConfig(),
			expected: false,
		},
		{
			name: "config with transforms",
			config: &config.TransformationConfig{
				RequestTransforms: []config.RequestTransformFunc{
					func(data model.UniData) (model.UniData, error) {
						return data, nil
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.HasRequestTransforms())
		})
	}
}

func TestTransformationConfig_HasResponseTransforms(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.TransformationConfig
		expected bool
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: false,
		},
		{
			name:     "empty config",
			config:   config.NewTransformationConfig(),
			expected: false,
		},
		{
			name: "config with transforms",
			config: &config.TransformationConfig{
				ResponseTransforms: []config.ResponseTransformFunc{
					func(data model.UniData) (model.UniData, error) {
						return data, nil
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.HasResponseTransforms())
		})
	}
}

func TestTransformationConfig_HasAnyTransforms(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.TransformationConfig
		expected bool
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: false,
		},
		{
			name:     "empty config",
			config:   config.NewTransformationConfig(),
			expected: false,
		},
		{
			name: "config with request transforms only",
			config: &config.TransformationConfig{
				RequestTransforms: []config.RequestTransformFunc{
					func(data model.UniData) (model.UniData, error) {
						return data, nil
					},
				},
			},
			expected: true,
		},
		{
			name: "config with response transforms only",
			config: &config.TransformationConfig{
				ResponseTransforms: []config.ResponseTransformFunc{
					func(data model.UniData) (model.UniData, error) {
						return data, nil
					},
				},
			},
			expected: true,
		},
		{
			name: "config with both transforms",
			config: &config.TransformationConfig{
				RequestTransforms: []config.RequestTransformFunc{
					func(data model.UniData) (model.UniData, error) {
						return data, nil
					},
				},
				ResponseTransforms: []config.ResponseTransformFunc{
					func(data model.UniData) (model.UniData, error) {
						return data, nil
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.HasAnyTransforms())
		})
	}
}

func TestRequestTransformFunc_Integration(t *testing.T) {
	// Test that RequestTransformFunc works with UniData
	originalData := model.UniData{
		Path:        "/test/123",
		IDs:         []string{"123"},
		ContentType: "application/json",
		Body:        []byte(`{"id": "123", "name": "test"}`),
	}

	transform := func(data model.UniData) (model.UniData, error) {
		if len(data.IDs) == 0 {
			return model.UniData{}, assert.AnError
		}
		// Modify the data
		modifiedData := data
		modifiedData.Body = []byte(`{"id": "123", "name": "test", "transformed": true, "section": "test-section"}`)
		return modifiedData, nil
	}

	transformedData, err := transform(originalData)

	require.NoError(t, err)
	assert.Contains(t, string(transformedData.Body), `"transformed": true`)
	assert.Contains(t, string(transformedData.Body), "test-section")
	assert.Equal(t, originalData.Path, transformedData.Path)
	assert.Equal(t, originalData.IDs, transformedData.IDs)
}

func TestResponseTransformFunc_Integration(t *testing.T) {
	// Test that ResponseTransformFunc works with UniData
	originalData := model.UniData{
		Path:        "/test/123",
		IDs:         []string{"123"},
		ContentType: "application/json",
		Body:        []byte(`{"id": "123", "name": "test"}`),
	}

	transform := func(data model.UniData) (model.UniData, error) {
		if len(data.IDs) == 0 {
			return model.UniData{}, assert.AnError
		}
		// Modify the response data
		modifiedData := data
		modifiedData.Body = []byte(`{"id": "123", "name": "test", ` +
			`"response_transformed": true, "section": "test-section"}`)
		modifiedData.ContentType = "application/json; charset=utf-8"
		return modifiedData, nil
	}

	transformedData, err := transform(originalData)

	require.NoError(t, err)
	assert.Contains(t, string(transformedData.Body), `"response_transformed": true`)
	assert.Contains(t, string(transformedData.Body), "test-section")
	assert.Equal(t, "application/json; charset=utf-8", transformedData.ContentType)
	assert.Equal(t, originalData.Path, transformedData.Path)
	assert.Equal(t, originalData.IDs, transformedData.IDs)
}

func TestSection_WithTransformations(t *testing.T) {
	// Test that Section can hold transformation configuration
	section := config.Section{
		PathPattern:     "/test/*",
		CaseSensitive:   true,
		Transformations: config.NewTransformationConfig(),
	}

	// Add some transformations
	section.Transformations.AddRequestTransform(
		func(data model.UniData) (model.UniData, error) {
			return data, nil
		})

	section.Transformations.AddResponseTransform(
		func(data model.UniData) (model.UniData, error) {
			return data, nil
		})

	assert.True(t, section.Transformations.HasAnyTransforms())
	assert.True(t, section.Transformations.HasRequestTransforms())
	assert.True(t, section.Transformations.HasResponseTransforms())
}