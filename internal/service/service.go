package service

import (
	"context"

	"github.com/bmcszk/unimock/pkg/model"
)

// MockService handles the core mock functionality
type MockService interface {
	// GetResource retrieves a resource by path and ID
	GetResource(ctx context.Context, path string, id string) (*model.MockData, error)
	// GetResourcesByPath retrieves all resources at a given path
	GetResourcesByPath(ctx context.Context, path string) ([]*model.MockData, error)
	// CreateResource creates a new resource
	CreateResource(ctx context.Context, path string, ids []string, data *model.MockData) error
	// UpdateResource updates an existing resource
	UpdateResource(ctx context.Context, path string, id string, data *model.MockData) error
	// DeleteResource removes a resource
	DeleteResource(ctx context.Context, path string, id string) error
}

// ScenarioService manages test scenarios
type ScenarioService interface {
	// GetScenarioByPath returns a scenario matching the given path and method
	GetScenarioByPath(ctx context.Context, path string, method string) *model.Scenario
	// ListScenarios returns all available scenarios
	ListScenarios(ctx context.Context) []*model.Scenario
	// GetScenario retrieves a scenario by UUID
	GetScenario(ctx context.Context, uuid string) (*model.Scenario, error)
	// CreateScenario creates a new scenario
	CreateScenario(ctx context.Context, scenario *model.Scenario) error
	// UpdateScenario updates an existing scenario
	UpdateScenario(ctx context.Context, uuid string, scenario *model.Scenario) error
	// DeleteScenario removes a scenario
	DeleteScenario(ctx context.Context, uuid string) error
}

// TechService handles technical operations
type TechService interface {
	// GetHealthStatus returns the health status of the service
	GetHealthStatus(ctx context.Context) map[string]any
	// GetMetrics returns metrics about the service
	GetMetrics(ctx context.Context) map[string]any
	// IncrementRequestCount increments the request counter
	IncrementRequestCount(ctx context.Context, path string)
}
