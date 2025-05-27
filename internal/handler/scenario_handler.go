package handler

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/bmcszk/unimock/internal/service"
	"github.com/bmcszk/unimock/pkg/model"
)

// ScenarioHandler handles endpoints for managing scenarios
type ScenarioHandler struct {
	prefix  string
	service service.ScenarioService
	logger  *slog.Logger
}

// NewScenarioHandler creates a new instance of ScenarioHandler
func NewScenarioHandler(service service.ScenarioService, logger *slog.Logger) *ScenarioHandler {
	return &ScenarioHandler{
		prefix:  "/_uni/scenarios",
		service: service,
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

func (h *ScenarioHandler) handleList(w http.ResponseWriter, r *http.Request) {
	// Get all scenarios from service
	scenarios := h.service.ListScenarios(r.Context())

	// Write response
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(scenarios)
	if err != nil {
		h.logger.Error("failed to encode scenarios", "error", err)
	}
}

func (h *ScenarioHandler) handleGet(w http.ResponseWriter, r *http.Request, uuid string) {
	// Get the scenario from service
	scenario, err := h.service.GetScenario(r.Context(), uuid)
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

	// Create the scenario
	if err := h.service.CreateScenario(r.Context(), &scenario); err != nil {
		h.logger.Error("failed to create scenario", "error", err, "uuid", scenario.UUID)
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, "Scenario already exists", http.StatusConflict)
		} else if strings.Contains(err.Error(), "invalid") { // Covers validation errors from service
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to create scenario", http.StatusInternalServerError)
		}
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
	if _, err = w.Write(scenarioJSON); err != nil {
		h.logger.Error("failed to write scenario response", "error", err)
	}
}

func (h *ScenarioHandler) handleUpdate(w http.ResponseWriter, r *http.Request, uuid string) {
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

	// Update the scenario
	if err := h.service.UpdateScenario(r.Context(), uuid, &scenario); err != nil {
		h.logger.Error("failed to update scenario", "error", err, "uuid", uuid)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Scenario not found", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "mismatch") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to update scenario", http.StatusInternalServerError)
		}
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(scenario)
	if err != nil {
		h.logger.Error("failed to encode scenario", "error", err)
	}
}

func (h *ScenarioHandler) handleDelete(w http.ResponseWriter, r *http.Request, uuid string) {
	// Delete the scenario
	if err := h.service.DeleteScenario(r.Context(), uuid); err != nil {
		h.logger.Error("failed to delete scenario", "error", err, "uuid", uuid)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Scenario not found", http.StatusNotFound)
		} else if strings.Contains(err.Error(), "invalid") { // e.g. invalid UUID format if service checked
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to delete scenario", http.StatusInternalServerError)
		}
		return
	}

	// Write response
	w.WriteHeader(http.StatusNoContent)
}
