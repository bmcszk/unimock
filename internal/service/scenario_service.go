package service

import (
	"context"
	"fmt"

	// "log/slog"
	// "os"
	"strings"

	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/model"
	"github.com/google/uuid"
)

// scenarioService implements the ScenarioService interface
type scenarioService struct {
	storage storage.ScenarioStorage
	// logger  *slog.Logger
}

// NewScenarioService creates a new instance of ScenarioService
func NewScenarioService(storage storage.ScenarioStorage) ScenarioService {
	// logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return &scenarioService{
		storage: storage,
		// logger:  logger,
	}
}

// GetScenarioByPath returns a scenario matching the given path and method
func (s *scenarioService) GetScenarioByPath(ctx context.Context, path string, method string) *model.Scenario {
	// s.logger.Debug("GetScenarioByPath called", "path", path, "method", method)
	scenarios := s.storage.List()
	// s.logger.Debug("GetScenarioByPath: scenarios from storage", "count", len(scenarios))
	for _, scenario := range scenarios {
		// s.logger.Debug("GetScenarioByPath: checking scenario", "index", i, "uuid", scenario.UUID, "requestPath", scenario.RequestPath)

		// Split the RequestPath into method and path
		parts := strings.SplitN(scenario.RequestPath, " ", 2)
		if len(parts) != 2 {
			// s.logger.Debug("GetScenarioByPath: scenario has invalid RequestPath format", "index", i, "requestPath", scenario.RequestPath)
			continue
		}

		scenarioMethod := parts[0]
		scenarioPath := parts[1]
		// s.logger.Debug("GetScenarioByPath: parsed scenario parts", "index", i, "scenarioMethod", scenarioMethod, "scenarioPath", scenarioPath)

		// Match method first
		if scenarioMethod != method {
			// s.logger.Debug("GetScenarioByPath: method mismatch", "index", i, "scenarioMethod", scenarioMethod, "expectedMethod", method)
			continue
		}

		// Check for exact path match
		if scenarioPath == path {
			// s.logger.Debug("GetScenarioByPath: exact path match found", "index", i, "uuid", scenario.UUID)
			return scenario
		}

		// Check for wildcard path match
		if strings.HasSuffix(scenarioPath, "/*") {
			basePath := strings.TrimSuffix(scenarioPath, "/*")
			// s.logger.Debug("GetScenarioByPath: checking wildcard match", "index", i, "basePath", basePath, "targetPath", path)
			if strings.HasPrefix(path, basePath+"/") {
				// s.logger.Debug("GetScenarioByPath: wildcard path match found", "index", i, "uuid", scenario.UUID)
				return scenario
			}
		}
		// s.logger.Debug("GetScenarioByPath: no match for scenario", "index", i, "uuid", scenario.UUID)
	}
	// s.logger.Debug("GetScenarioByPath: no scenario matched")
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
	// Validate scenario basic fields first
	if err := s.validateScenario(scenario); err != nil {
		return err
	}

	// Ensure UUID in path matches UUID in body. The UUID in path is authoritative.
	if scenario.UUID == "" {
		// If body UUID is empty, it's acceptable, we use the path UUID.
		scenario.UUID = uuid
	} else if uuid != scenario.UUID {
		return fmt.Errorf("invalid request: UUID in path (%s) does not match UUID in scenario body (%s)", uuid, scenario.UUID)
	}

	err := s.storage.Update(uuid, scenario)
	if err != nil {
		// Standardized error message for not found resources
		return fmt.Errorf("resource not found")
	}

	// Validate status code
	if scenario.StatusCode < 100 || scenario.StatusCode > 599 {
		return fmt.Errorf("invalid status code")
	}

	// Validate content type - ensure it is not empty if provided, but don't restrict to application/*
	// A more sophisticated validation (e.g. RFC media type format) could be added if needed.
	if scenario.ContentType == "" {
		// Allow empty ContentType if that's acceptable for some scenarios (e.g. 204 No Content)
		// If ContentType is mandatory, this should be an error.
		// For now, let's assume it's optional or can be empty for certain status codes.
	} else if strings.ContainsAny(scenario.ContentType, " \t\n\r") {
		// Basic check for obviously invalid characters, but very simplistic.
		return fmt.Errorf("invalid content type: contains whitespace characters")
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
		// If method was already validated and part of RequestPath, this check might be redundant here
		// but as a direct validation of scenario model, it's fine.
		return fmt.Errorf("invalid HTTP method in request path: %s", method)
	}

	// Validate status code
	if scenario.StatusCode < 100 || scenario.StatusCode > 599 {
		return fmt.Errorf("invalid status code: %d", scenario.StatusCode)
	}

	// Validate content type - ensure it is not empty if a body is expected or content is relevant.
	// For now, a very basic check: if ContentType is provided, it should not contain obvious invalid characters.
	// An empty ContentType might be valid for responses like 204 No Content or redirects.
	if scenario.ContentType != "" && strings.ContainsAny(scenario.ContentType, " \t\n\r") {
		return fmt.Errorf("invalid content type: contains whitespace characters")
	}

	// Removed overly restrictive check: !strings.HasPrefix(scenario.ContentType, "application/")

	return nil
}
