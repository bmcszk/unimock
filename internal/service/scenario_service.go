package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/model"
	"github.com/google/uuid"
)

// scenarioService implements the ScenarioService interface
type scenarioService struct {
	storage storage.ScenarioStorage
}

// NewScenarioService creates a new instance of ScenarioService
func NewScenarioService(storage storage.ScenarioStorage) ScenarioService {
	return &scenarioService{
		storage: storage,
	}
}

// GetScenarioByPath returns a scenario matching the given path and method
func (s *scenarioService) GetScenarioByPath(ctx context.Context, path string, method string) *model.Scenario {
	scenarios := s.storage.List()

	for _, scenario := range scenarios {
		// Split the RequestPath into method and path
		parts := strings.SplitN(scenario.RequestPath, " ", 2)
		if len(parts) != 2 {
			continue
		}

		scenarioMethod := parts[0]
		scenarioPath := parts[1]

		// Match method first
		if scenarioMethod != method {
			continue
		}

		// Check for exact path match
		if scenarioPath == path {
			return scenario
		}

		// Check for wildcard path match
		if strings.HasSuffix(scenarioPath, "/*") {
			basePath := strings.TrimSuffix(scenarioPath, "/*")
			if strings.HasPrefix(path, basePath+"/") {
				return scenario
			}
		}
	}
	return nil
}

// ListScenarios returns all available scenarios
func (s *scenarioService) ListScenarios(ctx context.Context) []*model.Scenario {
	return s.storage.List()
}

// GetScenario retrieves a scenario by UUID
func (s *scenarioService) GetScenario(ctx context.Context, uuid string) (*model.Scenario, error) {
	if uuid == "" {
		return nil, fmt.Errorf("invalid request: scenario ID cannot be empty")
	}
	scenario, err := s.storage.Get(uuid)
	if err != nil {
		// Return standardized error message regardless of the specific error from storage
		return nil, fmt.Errorf("resource not found")
	}
	return scenario, nil
}

// CreateScenario creates a new scenario
func (s *scenarioService) CreateScenario(ctx context.Context, scenario *model.Scenario) error {
	// Validate scenario
	if err := s.validateScenario(scenario); err != nil {
		return err
	}

	// Generate UUID if not provided
	if scenario.UUID == "" {
		scenario.UUID = uuid.New().String()
	}

	// Set location if not provided
	if scenario.Location == "" {
		scenario.Location = "/_uni/scenarios/" + scenario.UUID
	}

	err := s.storage.Create(scenario.UUID, scenario)
	if err != nil {
		// Standardized error message for already existing resources
		return fmt.Errorf("resource already exists")
	}
	return nil
}

// UpdateScenario updates an existing scenario
func (s *scenarioService) UpdateScenario(ctx context.Context, uuid string, scenario *model.Scenario) error {
	// Validate scenario
	if err := s.validateScenario(scenario); err != nil {
		return err
	}

	// Ensure UUID matches
	if uuid != scenario.UUID {
		return fmt.Errorf("resource not found")
	}

	err := s.storage.Update(uuid, scenario)
	if err != nil {
		// Standardized error message for not found resources
		return fmt.Errorf("resource not found")
	}
	return nil
}

// DeleteScenario removes a scenario
func (s *scenarioService) DeleteScenario(ctx context.Context, uuid string) error {
	if uuid == "" {
		return fmt.Errorf("invalid request: scenario ID cannot be empty")
	}
	err := s.storage.Delete(uuid)
	if err != nil {
		// Standardized error message for not found resources
		return fmt.Errorf("resource not found")
	}
	return nil
}

// validateScenario validates a scenario
func (s *scenarioService) validateScenario(scenario *model.Scenario) error {
	if scenario == nil {
		return fmt.Errorf("invalid request: scenario cannot be nil")
	}

	// Validate request path format
	parts := strings.SplitN(scenario.RequestPath, " ", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid request path format")
	}

	// Validate HTTP method
	method := parts[0]
	validMethods := map[string]bool{
		"GET":     true,
		"POST":    true,
		"PUT":     true,
		"DELETE":  true,
		"PATCH":   true,
		"HEAD":    true,
		"OPTIONS": true,
	}
	if !validMethods[method] {
		return fmt.Errorf("invalid request path format")
	}

	// Validate status code
	if scenario.StatusCode < 100 || scenario.StatusCode > 599 {
		return fmt.Errorf("invalid status code")
	}

	// Validate content type
	if !strings.HasPrefix(scenario.ContentType, "application/") {
		return fmt.Errorf("invalid content type")
	}

	return nil
}
