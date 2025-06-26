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
	splitParts       = 2
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
	var exactMatch *model.Scenario
	var wildcardMatch *model.Scenario

	for _, scenario := range scenarios {
		scenarioMethod, scenarioPath := s.parseRequestPath(scenario.RequestPath)
		if scenarioMethod == "" || scenarioMethod != method {
			continue
		}

		exactMatch = s.checkExactMatch(&scenario, exactMatch, scenarioPath, path)
		wildcardMatch = s.checkWildcardMatch(&scenario, wildcardMatch, scenarioPath, path)
	}

	if match := s.selectBestMatch(exactMatch, wildcardMatch); match != nil {
		return *match, true
	}
	return model.Scenario{}, false
}

// parseRequestPath extracts method and path from scenario request path
func (*ScenarioService) parseRequestPath(requestPath string) (method, path string) {
	parts := strings.SplitN(requestPath, " ", requestPathParts)
	if len(parts) != requestPathParts {
		return "", ""
	}
	return parts[0], parts[singleItem]
}

// checkExactMatch checks if scenario matches exact path and updates exactMatch if needed
func (*ScenarioService) checkExactMatch(
	scenario, exactMatch *model.Scenario, scenarioPath, path string,
) *model.Scenario {
	if scenarioPath == path && exactMatch == nil {
		return scenario
	}
	return exactMatch
}

// checkWildcardMatch checks if scenario matches wildcard path and updates wildcardMatch if needed
func (s *ScenarioService) checkWildcardMatch(
	scenario, wildcardMatch *model.Scenario, scenarioPath, path string,
) *model.Scenario {
	if strings.HasSuffix(scenarioPath, "/*") {
		return s.handleWildcardMatch(scenario, wildcardMatch, scenarioPath, path)
	}
	return wildcardMatch
}

// selectBestMatch returns the best match, preferring exact over wildcard
func (*ScenarioService) selectBestMatch(exactMatch, wildcardMatch *model.Scenario) *model.Scenario {
	if exactMatch != nil {
		return exactMatch
	}
	return wildcardMatch
}

// handleWildcardMatch processes wildcard scenario matching
func (s *ScenarioService) handleWildcardMatch(
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
func (*ScenarioService) selectBestWildcardMatch(
	scenario, wildcardMatch *model.Scenario, newMatchBase string,
) *model.Scenario {
	currentWildcardBase := strings.TrimSuffix(strings.SplitN(wildcardMatch.RequestPath, " ", 2)[1], "/*")
	if len(newMatchBase) > len(currentWildcardBase) {
		return scenario
	}
	return wildcardMatch
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
	if err := s.validateScenario(&scenario); err != nil {
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
	if err := s.validateScenario(&scenario); err != nil {
		return err
	}

	// Validate and set UUID consistency
	if err := s.validateUUIDConsistency(id, &scenario); err != nil {
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
func (*ScenarioService) validateUUIDConsistency(id string, scenario *model.Scenario) error {
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
func (s *ScenarioService) performStorageUpdate(id string, scenario model.Scenario) error {
	err := s.storage.Update(id, scenario)
	if err != nil {
		return errors.New("resource not found")
	}
	return nil
}

// validatePostUpdate performs additional validations after storage update
func (*ScenarioService) validatePostUpdate(_ model.Scenario) error {
	// These validations are already done in validateScenario, so this is just for extra safety
	// In case there were any modifications during the update process
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
func (*ScenarioService) validateScenario(scenario *model.Scenario) error {
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

	return nil
}
