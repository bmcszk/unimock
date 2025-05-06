package storage

import (
	"sync"

	"github.com/bmcszk/unimock/internal/errors"
	"github.com/bmcszk/unimock/internal/model"
)

// ScenarioStorage manages scenarios
type ScenarioStorage interface {
	Create(id string, scenario *model.Scenario) error
	Get(id string) (*model.Scenario, error)
	Update(id string, scenario *model.Scenario) error
	Delete(id string) error
	List() []*model.Scenario
}

// scenarioStorage implements the ScenarioStorage interface
type scenarioStorage struct {
	mu        *sync.RWMutex
	scenarios map[string]*model.Scenario
}

// NewScenarioStorage creates a new instance of ScenarioStorage
func NewScenarioStorage() ScenarioStorage {
	return &scenarioStorage{
		mu:        &sync.RWMutex{},
		scenarios: make(map[string]*model.Scenario),
	}
}

func (s *scenarioStorage) Create(id string, scenario *model.Scenario) error {
	if id == "" {
		return errors.NewInvalidRequestError("scenario ID cannot be empty")
	}
	if scenario == nil {
		return errors.NewInvalidRequestError("scenario cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if scenario already exists
	if _, exists := s.scenarios[id]; exists {
		return errors.NewConflictError(id)
	}

	// Store the scenario
	s.scenarios[id] = scenario

	return nil
}

func (s *scenarioStorage) Get(id string) (*model.Scenario, error) {
	if id == "" {
		return nil, errors.NewInvalidRequestError("scenario ID cannot be empty")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	scenario, exists := s.scenarios[id]
	if !exists {
		return nil, errors.NewNotFoundError(id, "")
	}

	return scenario, nil
}

func (s *scenarioStorage) Update(id string, scenario *model.Scenario) error {
	if id == "" {
		return errors.NewInvalidRequestError("scenario ID cannot be empty")
	}
	if scenario == nil {
		return errors.NewInvalidRequestError("scenario cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if scenario exists
	if _, exists := s.scenarios[id]; !exists {
		return errors.NewNotFoundError(id, "")
	}

	// Update the scenario
	s.scenarios[id] = scenario

	return nil
}

func (s *scenarioStorage) Delete(id string) error {
	if id == "" {
		return errors.NewInvalidRequestError("scenario ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if scenario exists
	if _, exists := s.scenarios[id]; !exists {
		return errors.NewNotFoundError(id, "")
	}

	// Remove scenario
	delete(s.scenarios, id)

	return nil
}

func (s *scenarioStorage) List() []*model.Scenario {
	s.mu.RLock()
	defer s.mu.RUnlock()

	scenarios := make([]*model.Scenario, 0, len(s.scenarios))
	for _, scenario := range s.scenarios {
		scenarios = append(scenarios, scenario)
	}

	return scenarios
}
