package handler_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/bmcszk/unimock/internal/handler"
	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create test handler with transformations
func createHandlerWithTransforms(transformConfig *config.TransformationConfig) *handler.UniHandler {
	store := storage.NewUniStorage()
	scenarioStore := storage.NewScenarioStorage()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	section := config.Section{
		PathPattern:     "/users/*",
		BodyIDPaths:     []string{"/id"},
		CaseSensitive:   false,
		Transformations: transformConfig,
	}

	mockConfig := &config.UniConfig{
		Sections: map[string]config.Section{
			"users": section,
		},
	}

	uniService := service.NewUniService(store, mockConfig)
	scenarioService := service.NewScenarioService(scenarioStore)
	return handler.NewUniHandler(uniService, scenarioService, logger, mockConfig)
}

func TestTransformationSimple_ResponseHeaders(t *testing.T) {
	// Create transformation config that modifies response data
	transformConfig := config.NewTransformationConfig()
	transformConfig.AddResponseTransform(
		func(data model.UniData) (model.UniData, error) {
			// Add transformation metadata to the response body
			modifiedData := data
			modifiedData.Body = []byte(`{"id": "123", "name": "test user", ` +
				`"x-transformed": "true", "x-section": "users"}`)
			return modifiedData, nil
		})

	_ = createHandlerWithTransforms(transformConfig)

	// Create test data
	ctx := context.Background()
	store := storage.NewUniStorage()
	uniService := service.NewUniService(store, &config.UniConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern:   "/users/*",
				BodyIDPaths:   []string{"/id"},
				CaseSensitive: false,
			},
		},
	})

	testData := model.UniData{
		Body:        []byte(`{"id": "123", "name": "test user"}`),
		ContentType: "application/json",
	}
	err := uniService.CreateResource(ctx, "users", false, []string{"123"}, testData)
	require.NoError(t, err)

	// Copy test data to the handler's service storage
	// Note: In a real scenario, we'd have a shared storage instance
	handlerStore := storage.NewUniStorage()
	handlerUniService := service.NewUniService(handlerStore, &config.UniConfig{
		Sections: map[string]config.Section{
			"users": {
				PathPattern:   "/users/*",
				BodyIDPaths:   []string{"/id"},
				CaseSensitive: false,
			},
		},
	})
	handlerTestData := model.UniData{
		Body:        []byte(`{"id": "123", "name": "test user"}`),
		ContentType: "application/json",
	}
	err = handlerUniService.CreateResource(ctx, "users", false, []string{"123"}, handlerTestData)
	require.NoError(t, err)

	// Re-create handler with the correct storage
	scenarioStore := storage.NewScenarioStorage()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	section := config.Section{
		PathPattern:     "/users/*",
		BodyIDPaths:     []string{"/id"},
		CaseSensitive:   false,
		Transformations: transformConfig,
	}

	mockConfig := &config.UniConfig{
		Sections: map[string]config.Section{
			"users": section,
		},
	}

	scenarioService := service.NewScenarioService(scenarioStore)
	uniHandler := handler.NewUniHandler(handlerUniService, scenarioService, logger, mockConfig)

	// Test request
	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	resp, err := uniHandler.HandleRequest(context.Background(), req)

	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify response transformations were applied
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), `"x-transformed": "true"`)
	assert.Contains(t, string(body), `"x-section": "users"`)
	assert.Contains(t, string(body), "test user")
}

func TestTransformationSimple_NoTransformations(t *testing.T) {
	// Create handler without transformations
	uniHandler := createHandlerWithTransforms(nil)

	// Test that it works normally
	req := httptest.NewRequest(http.MethodGet, "/users/nonexistent", nil)
	resp, err := uniHandler.HandleRequest(context.Background(), req)

	require.NoError(t, err)
	defer resp.Body.Close()
	// Should return 404 for non-existent resource
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestTransformationSimple_ErrorHandling(t *testing.T) {
	// Create transformation config that always fails
	transformConfig := config.NewTransformationConfig()
	transformConfig.AddRequestTransform(
		func(_ model.UniData) (model.UniData, error) {
			return model.UniData{}, assert.AnError // Use a known error from testify
		})

	uniHandler := createHandlerWithTransforms(transformConfig)

	// Use POST request to trigger request transformation (which will fail)
	reqBody := strings.NewReader(`{"id": "123", "name": "test user"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", reqBody)
	req.Header.Set("Content-Type", "application/json")
	resp, err := uniHandler.HandleRequest(context.Background(), req)

	require.NoError(t, err) // Handler returns response, not error
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode) // Now returns 500 for all transformation errors
}