package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bmcszk/unimock/internal/model"
	"github.com/bmcszk/unimock/internal/storage"

	"github.com/google/uuid"
)

// ScenarioHandler handles endpoints for managing scenarios
type ScenarioHandler struct {
	prefix  string
	storage storage.ScenarioStorage
	logger  *slog.Logger
}

// NewScenarioHandler creates a new instance of ScenarioHandler
func NewScenarioHandler(storage storage.ScenarioStorage, logger *slog.Logger) *ScenarioHandler {
	return &ScenarioHandler{
		prefix:  "/_uni/scenarios",
		storage: storage,
		logger:  logger,
	}
}

// GetScenarioByPath returns a scenario matching the given path, or nil if not found
func (h *ScenarioHandler) GetScenarioByPath(path string) *model.Scenario {
	scenarios := h.storage.List()
	for _, scenario := range scenarios {
		if scenario.Path == path {
			return scenario
		}
	}
	return nil
}

// ServeHTTP implements the http.Handler interface
func (h *ScenarioHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Log the request
	h.logger.Info("scenario endpoint request",
		"method", r.Method,
		"path", r.URL.Path)

	// Get the path without prefix
	path := strings.TrimPrefix(r.URL.Path, h.prefix)

	// Handle based on method and path
	switch {
	case r.Method == http.MethodGet && path == "":
		// GET /_uni/scenarios - List all scenarios
		h.handleList(w, r)
	case r.Method == http.MethodGet && path != "":
		// GET /_uni/scenarios/{uuid} - Get a specific scenario
		uuid := strings.TrimPrefix(path, "/")
		h.handleGet(w, r, uuid)
	case r.Method == http.MethodPost && path == "":
		// POST /_uni/scenarios - Create a new scenario
		h.handleCreate(w, r)
	case r.Method == http.MethodPut && path != "":
		// PUT /_uni/scenarios/{uuid} - Update a scenario
		uuid := strings.TrimPrefix(path, "/")
		h.handleUpdate(w, r, uuid)
	case r.Method == http.MethodDelete && path != "":
		// DELETE /_uni/scenarios/{uuid} - Delete a scenario
		uuid := strings.TrimPrefix(path, "/")
		h.handleDelete(w, r, uuid)
	default:
		http.NotFound(w, r)
	}
}

func (h *ScenarioHandler) handleList(w http.ResponseWriter, r *http.Request) {
	// Get all scenarios from storage
	scenarios := h.storage.List()

	// Write response
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(scenarios)
	if err != nil {
		h.logger.Error("failed to encode scenarios", "error", err)
	}
}

func (h *ScenarioHandler) handleGet(w http.ResponseWriter, r *http.Request, uuid string) {
	// Get the scenario from storage
	scenario, err := h.storage.Get(uuid)
	if err != nil {
		h.logger.Error("failed to get scenario", "error", err, "uuid", uuid)
		http.Error(w, "Scenario not found", http.StatusNotFound)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(scenario)
	if err != nil {
		h.logger.Error("failed to encode scenario", "error", err)
	}
}

func (h *ScenarioHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read request body", "error", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Parse the scenario
	var scenario model.Scenario
	if err := json.Unmarshal(body, &scenario); err != nil {
		h.logger.Error("failed to unmarshal scenario", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Generate UUID if not provided
	if scenario.UUID == "" {
		scenario.UUID = uuid.New().String()
	}

	// Set location if not provided
	if scenario.Location == "" {
		scenario.Location = h.prefix + "/" + scenario.UUID
	}

	// Store the scenario
	if err := h.storage.Create(scenario.UUID, &scenario); err != nil {
		h.logger.Error("failed to create scenario", "error", err, "uuid", scenario.UUID)
		http.Error(w, "Failed to create scenario", http.StatusConflict)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", scenario.Location)
	w.WriteHeader(http.StatusCreated)

	// Convert scenario to JSON and write response
	scenarioJSON, err := json.Marshal(scenario)
	if err != nil {
		h.logger.Error("failed to marshal scenario", "error", err)
		return
	}
	w.Write(scenarioJSON)
}

func (h *ScenarioHandler) handleUpdate(w http.ResponseWriter, r *http.Request, uuid string) {
	// Check if the scenario exists
	_, err := h.storage.Get(uuid)
	if err != nil {
		h.logger.Error("scenario not found", "error", err, "uuid", uuid)
		http.Error(w, "Scenario not found", http.StatusNotFound)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read request body", "error", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Parse the scenario
	var scenario model.Scenario
	if err := json.Unmarshal(body, &scenario); err != nil {
		h.logger.Error("failed to unmarshal scenario", "error", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Ensure UUID matches
	scenario.UUID = uuid

	// Update location if not provided
	if scenario.Location == "" {
		scenario.Location = h.prefix + "/" + scenario.UUID
	}

	// Update the scenario
	if err := h.storage.Update(uuid, &scenario); err != nil {
		h.logger.Error("failed to update scenario", "error", err, "uuid", uuid)
		http.Error(w, "Failed to update scenario", http.StatusInternalServerError)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")

	// Convert scenario to JSON and write response
	scenarioJSON, err := json.Marshal(scenario)
	if err != nil {
		h.logger.Error("failed to marshal scenario", "error", err)
		return
	}
	w.Write(scenarioJSON)
}

func (h *ScenarioHandler) handleDelete(w http.ResponseWriter, r *http.Request, uuid string) {
	// Check if the scenario exists
	_, err := h.storage.Get(uuid)
	if err != nil {
		h.logger.Error("scenario not found", "error", err, "uuid", uuid)
		http.Error(w, "Scenario not found", http.StatusNotFound)
		return
	}

	// Delete the scenario
	if err := h.storage.Delete(uuid); err != nil {
		h.logger.Error("failed to delete scenario", "error", err, "uuid", uuid)
		http.Error(w, "Failed to delete scenario", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}
