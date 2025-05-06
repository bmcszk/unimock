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
	storage storage.Storage
	logger  *slog.Logger
}

// NewScenarioHandler creates a new instance of ScenarioHandler
func NewScenarioHandler(storage storage.Storage, logger *slog.Logger) *ScenarioHandler {
	return &ScenarioHandler{
		prefix:  "/_uni/scenarios",
		storage: storage,
		logger:  logger,
	}
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

// handleList handles GET requests to list all scenarios
func (h *ScenarioHandler) handleList(w http.ResponseWriter, r *http.Request) {
	scenarios := []model.Scenario{}

	err := h.storage.ForEach(func(id string, data *model.MockData) error {
		// Only include data stored by this handler (path starting with our prefix)
		if strings.HasPrefix(data.Path, h.prefix) {
			var scenario model.Scenario
			if err := json.Unmarshal(data.Body, &scenario); err != nil {
				return err
			}
			scenarios = append(scenarios, scenario)
		}
		return nil
	})

	if err != nil {
		h.logger.Error("failed to list scenarios", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(scenarios)
	if err != nil {
		h.logger.Error("failed to encode scenarios", "error", err)
	}
}

// handleGet handles GET requests to retrieve a specific scenario
func (h *ScenarioHandler) handleGet(w http.ResponseWriter, r *http.Request, uuid string) {
	// Get the scenario from storage
	data, err := h.storage.Get(uuid)
	if err != nil {
		h.logger.Error("failed to get scenario", "error", err, "uuid", uuid)
		http.Error(w, "Scenario not found", http.StatusNotFound)
		return
	}

	// Parse the scenario
	var scenario model.Scenario
	if err := json.Unmarshal(data.Body, &scenario); err != nil {
		h.logger.Error("failed to unmarshal scenario", "error", err, "uuid", uuid)
		http.Error(w, "Invalid scenario data", http.StatusInternalServerError)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(scenario)
	if err != nil {
		h.logger.Error("failed to encode scenario", "error", err)
	}
}

// handleCreate handles POST requests to create a new scenario
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

	// Convert scenario to JSON
	scenarioJSON, err := json.Marshal(scenario)
	if err != nil {
		h.logger.Error("failed to marshal scenario", "error", err)
		http.Error(w, "Failed to create scenario", http.StatusInternalServerError)
		return
	}

	// Create mock data for storage
	data := &model.MockData{
		Path:        h.prefix,
		Location:    scenario.Location,
		ContentType: "application/json",
		Body:        scenarioJSON,
	}

	// Store the scenario
	if err := h.storage.Create([]string{scenario.UUID}, data); err != nil {
		h.logger.Error("failed to create scenario", "error", err, "uuid", scenario.UUID)
		http.Error(w, "Failed to create scenario", http.StatusConflict)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", scenario.Location)
	w.WriteHeader(http.StatusCreated)
	w.Write(scenarioJSON)
}

// handleUpdate handles PUT requests to update a scenario
func (h *ScenarioHandler) handleUpdate(w http.ResponseWriter, r *http.Request, uuid string) {
	// Check if the scenario exists
	existingData, err := h.storage.Get(uuid)
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

	// Convert scenario to JSON
	scenarioJSON, err := json.Marshal(scenario)
	if err != nil {
		h.logger.Error("failed to marshal scenario", "error", err)
		http.Error(w, "Failed to update scenario", http.StatusInternalServerError)
		return
	}

	// Update mock data for storage
	existingData.Body = scenarioJSON
	existingData.Location = scenario.Location

	// Update the scenario
	if err := h.storage.Update(uuid, existingData); err != nil {
		h.logger.Error("failed to update scenario", "error", err, "uuid", uuid)
		http.Error(w, "Failed to update scenario", http.StatusInternalServerError)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.Write(scenarioJSON)
}

// handleDelete handles DELETE requests to delete a scenario
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
