package service

import (
	"context"
	"errors"
	"fmt"

	// "log/slog"
	// "os"
	"strings"

	"github.com/bmcszk/unimock/internal/storage"
	"github.com/bmcszk/unimock/pkg/model"
	"github.com/google/uuid"
)

const (
	requestPathParts = 2
	singleItem       = 1
	minStatusCode    = 100
	maxStatusCode    = 599
)

// ScenarioService manages test scenarios
type ScenarioService struct {
	storage storage.ScenarioStorage
}

// NewScenarioService creates a new instance of ScenarioService
func NewScenarioService(scenarioStorage storage.ScenarioStorage) *ScenarioService {
	return &ScenarioService{
		storage: scenarioStorage,
	}
}

// GetScenarioByPath is a convenience method primarily for testing.
// It iterates through scenarios to find a match based on method and path (exact or wildcard).
func (s *ScenarioService) GetScenarioByPath(_ context.Context, path string, method string) (model.Scenario, bool) {
	scenarios := s.storage.List()
	return s.findBestScenarioMatch(scenarios, path, method)
}

// findBestScenarioMatch searches through scenarios to find the best match
func (s *ScenarioService) findBestScenarioMatch(
	scenarios []model.Scenario, path, method string,
) (model.Scenario, bool) {
	var exactMatch model.Scenario
	var wildcardMatch model.Scenario

	for _, scenario := range scenarios {
		if !s.isMethodMatch(scenario, method) {
			continue
		}

		if s.tryExactMatch(&exactMatch, scenario, path) {
			continue
		}

		s.tryWildcardMatch(&wildcardMatch, scenario, path)
	}

	return s.selectBestMatch(exactMatch, wildcardMatch)
}

// isMethodMatch checks if scenario matches the HTTP method
func (s *ScenarioService) isMethodMatch(scenario model.Scenario, method string) bool {
	scenarioMethod, _ := s.parseRequestPath(scenario.RequestPath)
	return scenarioMethod != "" && scenarioMethod == method
}

// tryExactMatch attempts to find an exact match
func (s *ScenarioService) tryExactMatch(exactMatch *model.Scenario, scenario model.Scenario, path string) bool {
	if exactMatch.UUID != "" {
		return false // Already have exact match
	}

	_, scenarioPath := s.parseRequestPath(scenario.RequestPath)
	if match, found := s.checkExactMatch(scenario, scenarioPath, path); found {
		*exactMatch = match
		return true // Found exact match
	}
	return false
}

// tryWildcardMatch attempts to find a wildcard match
func (s *ScenarioService) tryWildcardMatch(wildcardMatch *model.Scenario, scenario model.Scenario, path string) bool {
	if wildcardMatch.UUID != "" {
		return false // Already have wildcard match
	}

	_, scenarioPath := s.parseRequestPath(scenario.RequestPath)
	if match, found := s.checkWildcardMatch(scenario, scenarioPath, path); found {
		*wildcardMatch = match
		return true
	}
	return false
}

// parseRequestPath extracts method and path from scenario request path
func (*ScenarioService) parseRequestPath(requestPath string) (method, path string) {
	parts := strings.SplitN(requestPath, " ", requestPathParts)
	if len(parts) != requestPathParts {
		return "", ""
	}
	return parts[0], parts[singleItem]
}

// checkExactMatch checks if scenario matches exact path and returns the match if found
func (*ScenarioService) checkExactMatch(
	scenario model.Scenario, scenarioPath, path string,
) (model.Scenario, bool) {
	if scenarioPath == path {
		return scenario, true
	}
	return model.Scenario{}, false
}

// checkWildcardMatch checks if scenario matches wildcard path and returns the match if found
func (s *ScenarioService) checkWildcardMatch(
	scenario model.Scenario, scenarioPath, path string,
) (model.Scenario, bool) {
	if strings.HasSuffix(scenarioPath, "/*") {
		return s.handleWildcardMatch(scenario, scenarioPath, path)
	}
	return model.Scenario{}, false
}

// selectBestMatch returns the best match, preferring exact over wildcard
func (*ScenarioService) selectBestMatch(
	exactMatch model.Scenario, wildcardMatch model.Scenario,
) (model.Scenario, bool) {
	if exactMatch.UUID != "" {
		return exactMatch, true
	}
	if wildcardMatch.UUID != "" {
		return wildcardMatch, true
	}
	return model.Scenario{}, false
}

// handleWildcardMatch processes wildcard scenario matching and returns the best match
func (*ScenarioService) handleWildcardMatch(
	scenario model.Scenario, scenarioPath, path string,
) (model.Scenario, bool) {
	basePath := strings.TrimSuffix(scenarioPath, "/*")
	if !strings.HasPrefix(path, basePath+"/") && path != basePath {
		return model.Scenario{}, false
	}

	return scenario, true
}

// ListScenarios returns all available scenarios
func (s *ScenarioService) ListScenarios(_ context.Context) []model.Scenario {
	return s.storage.List()
}

// GetScenario retrieves a scenario by UUID
func (s *ScenarioService) GetScenario(_ context.Context, id string) (model.Scenario, error) {
	if id == "" {
		return model.Scenario{}, errors.New("invalid request: scenario ID cannot be empty")
	}
	scenario, err := s.storage.Get(id)
	if err != nil {
		// Return standardized error message regardless of the specific error from storage
		return model.Scenario{}, errors.New("resource not found")
	}
	return scenario, nil
}

// CreateScenario creates a new scenario
func (s *ScenarioService) CreateScenario(_ context.Context, scenario model.Scenario) (model.Scenario, error) {
	// Validate scenario
	if err := s.validateScenario(scenario); err != nil {
		return model.Scenario{}, err
	}

	// Generate UUID if not provided
	if scenario.UUID == "" {
		scenario.UUID = uuid.New().String()
	}

	err := s.storage.Create(scenario.UUID, scenario)
	if err != nil {
		// Standardized error message for already existing resources
		return model.Scenario{}, errors.New("resource already exists")
	}
	return scenario, nil
}

// UpdateScenario updates an existing scenario
func (s *ScenarioService) UpdateScenario(_ context.Context, id string, scenario model.Scenario) error {
	// Validate scenario basic fields first
	if err := s.validateScenario(scenario); err != nil {
		return err
	}

	// Validate and set UUID consistency
	updatedScenario, err := s.validateUUIDConsistency(id, scenario)
	if err != nil {
		return err
	}
	scenario = updatedScenario

	// Perform the storage update
	return s.performStorageUpdate(id, scenario)
}

// validateUUIDConsistency ensures UUID in path matches UUID in scenario body
func (*ScenarioService) validateUUIDConsistency(id string, scenario model.Scenario) (model.Scenario, error) {
	if scenario.UUID == "" {
		scenario.UUID = id
		return scenario, nil
	}
	if id != scenario.UUID {
		return model.Scenario{}, fmt.Errorf(
			"invalid request: UUID in path (%s) does not match UUID in scenario body (%s)",
			id, scenario.UUID,
		)
	}
	return scenario, nil
}

// performStorageUpdate performs the actual storage update operation
func (s *ScenarioService) performStorageUpdate(id string, scenario model.Scenario) error {
	err := s.storage.Update(id, scenario)
	if err != nil {
		return errors.New("resource not found")
	}
	return nil
}

// DeleteScenario removes a scenario
func (s *ScenarioService) DeleteScenario(_ context.Context, id string) error {
	if id == "" {
		return errors.New("invalid request: scenario ID cannot be empty")
	}
	err := s.storage.Delete(id)
	if err != nil {
		// Standardized error message for not found resources
		return errors.New("resource not found")
	}
	return nil
}

// validateScenario validates a scenario
func (*ScenarioService) validateScenario(scenario model.Scenario) error {
	// Validate request path format
	parts := strings.SplitN(scenario.RequestPath, " ", 2)
	if len(parts) != 2 {
		return errors.New("invalid request path format")
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

	return nil
}
