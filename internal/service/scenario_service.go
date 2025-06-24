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
	singleItem = 1
	minStatusCode = 100
	maxStatusCode = 599
	splitParts = 2
)

// scenarioService implements the ScenarioService interface
// The ScenarioService interface is defined in service.go
type scenarioService struct {
	storage storage.ScenarioStorage
}

// NewScenarioService creates a new instance of ScenarioService
func NewScenarioService(scenarioStorage storage.ScenarioStorage) ScenarioService {
	return &scenarioService{
		storage: scenarioStorage,
	}
}

// GetScenarioByPath is a convenience method primarily for testing.
// It iterates through scenarios to find a match based on method and path (exact or wildcard).
func (s *scenarioService) GetScenarioByPath(_ context.Context, path string, method string) *model.Scenario {
	scenarios := s.storage.List()
	var exactMatch *model.Scenario
	var wildcardMatch *model.Scenario

	for _, scenario := range scenarios {
		scenarioMethod, scenarioPath := s.parseRequestPath(scenario.RequestPath)
		if scenarioMethod == "" || scenarioMethod != method {
			continue
		}

		exactMatch = s.checkExactMatch(scenario, exactMatch, scenarioPath, path)
		wildcardMatch = s.checkWildcardMatch(scenario, wildcardMatch, scenarioPath, path)
	}

	return s.selectBestMatch(exactMatch, wildcardMatch)
}

// parseRequestPath extracts method and path from scenario request path
func (*scenarioService) parseRequestPath(requestPath string) (method, path string) {
	parts := strings.SplitN(requestPath, " ", requestPathParts)
	if len(parts) != requestPathParts {
		return "", ""
	}
	return parts[0], parts[singleItem]
}

// checkExactMatch checks if scenario matches exact path and updates exactMatch if needed
func (*scenarioService) checkExactMatch(
	scenario, exactMatch *model.Scenario, scenarioPath, path string,
) *model.Scenario {
	if scenarioPath == path && exactMatch == nil {
		return scenario
	}
	return exactMatch
}

// checkWildcardMatch checks if scenario matches wildcard path and updates wildcardMatch if needed
func (s *scenarioService) checkWildcardMatch(
	scenario, wildcardMatch *model.Scenario, scenarioPath, path string,
) *model.Scenario {
	if strings.HasSuffix(scenarioPath, "/*") {
		return s.handleWildcardMatch(scenario, wildcardMatch, scenarioPath, path)
	}
	return wildcardMatch
}

// selectBestMatch returns the best match, preferring exact over wildcard
func (*scenarioService) selectBestMatch(exactMatch, wildcardMatch *model.Scenario) *model.Scenario {
	if exactMatch != nil {
		return exactMatch
	}
	return wildcardMatch
}

// handleWildcardMatch processes wildcard scenario matching
func (s *scenarioService) handleWildcardMatch(
	scenario, wildcardMatch *model.Scenario, scenarioPath, path string,
) *model.Scenario {
	basePath := strings.TrimSuffix(scenarioPath, "/*")
	if !strings.HasPrefix(path, basePath+"/") && path != basePath {
		return wildcardMatch
	}

	if wildcardMatch == nil {
		return scenario
	}

	if scenario == wildcardMatch {
		return wildcardMatch
	}

	return s.selectBestWildcardMatch(scenario, wildcardMatch, basePath)
}

// selectBestWildcardMatch chooses the most specific wildcard match
func (*scenarioService) selectBestWildcardMatch(
	scenario, wildcardMatch *model.Scenario, newMatchBase string,
) *model.Scenario {
	currentWildcardBase := strings.TrimSuffix(strings.SplitN(wildcardMatch.RequestPath, " ", 2)[1], "/*")
	if len(newMatchBase) > len(currentWildcardBase) {
		return scenario
	}
	return wildcardMatch
}

// ListScenarios returns all available scenarios
func (s *scenarioService) ListScenarios(_ context.Context) []*model.Scenario {
	return s.storage.List()
}

// GetScenario retrieves a scenario by UUID
func (s *scenarioService) GetScenario(_ context.Context, id string) (*model.Scenario, error) {
	if id == "" {
		return nil, errors.New("invalid request: scenario ID cannot be empty")
	}
	scenario, err := s.storage.Get(id)
	if err != nil {
		// Return standardized error message regardless of the specific error from storage
		return nil, errors.New("resource not found")
	}
	return scenario, nil
}

// CreateScenario creates a new scenario
func (s *scenarioService) CreateScenario(_ context.Context, scenario *model.Scenario) error {
	// Validate scenario
	if err := s.validateScenario(scenario); err != nil {
		return err
	}

	// Generate UUID if not provided
	if scenario.UUID == "" {
		scenario.UUID = uuid.New().String()
	}

	err := s.storage.Create(scenario.UUID, scenario)
	if err != nil {
		// Standardized error message for already existing resources
		return errors.New("resource already exists")
	}
	return nil
}

// UpdateScenario updates an existing scenario
func (s *scenarioService) UpdateScenario(_ context.Context, id string, scenario *model.Scenario) error {
	// Validate scenario basic fields first
	if err := s.validateScenario(scenario); err != nil {
		return err
	}

	// Validate and set UUID consistency
	if err := s.validateUUIDConsistency(id, scenario); err != nil {
		return err
	}

	// Perform the storage update
	if err := s.performStorageUpdate(id, scenario); err != nil {
		return err
	}

	// Additional validations after storage update
	return s.validatePostUpdate(scenario)
}

// validateUUIDConsistency ensures UUID in path matches UUID in scenario body
func (*scenarioService) validateUUIDConsistency(id string, scenario *model.Scenario) error {
	if scenario.UUID == "" {
		scenario.UUID = id
		return nil
	}
	if id != scenario.UUID {
		return fmt.Errorf(
			"invalid request: UUID in path (%s) does not match UUID in scenario body (%s)",
			id, scenario.UUID,
		)
	}
	return nil
}

// performStorageUpdate performs the actual storage update operation
func (s *scenarioService) performStorageUpdate(id string, scenario *model.Scenario) error {
	err := s.storage.Update(id, scenario)
	if err != nil {
		return errors.New("resource not found")
	}
	return nil
}

// validatePostUpdate performs additional validations after storage update
func (*scenarioService) validatePostUpdate(_ *model.Scenario) error {
	// These validations are already done in validateScenario, so this is just for extra safety
	// In case there were any modifications during the update process
	return nil
}

// DeleteScenario removes a scenario
func (s *scenarioService) DeleteScenario(_ context.Context, id string) error {
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
func (*scenarioService) validateScenario(scenario *model.Scenario) error {
	if scenario == nil {
		return errors.New("invalid request: scenario cannot be nil")
	}

	// Validate request path format
	parts := strings.SplitN(scenario.RequestPath, " ", splitParts)
	if len(parts) != splitParts {
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

	// Validate status code
	if scenario.StatusCode < minStatusCode || scenario.StatusCode > maxStatusCode {
		return fmt.Errorf("invalid status code: %d", scenario.StatusCode)
	}

	// Validate content type - ensure it is not empty if a body is expected or content is relevant.
	// For now, a very basic check: if ContentType is provided, it should not contain obvious invalid characters.
	// An empty ContentType might be valid for responses like 204 No Content or redirects.
	if scenario.ContentType != "" && strings.ContainsAny(scenario.ContentType, " \t\n\r") {
		return errors.New("invalid content type: contains whitespace characters")
	}

	// Removed overly restrictive check: !strings.HasPrefix(scenario.ContentType, "application/")

	return nil
}
