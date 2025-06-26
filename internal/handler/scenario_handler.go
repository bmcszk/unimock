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

const (
	uuidLogKey = "uuid"
	applicationJSON = "application/json"
)

// ScenarioHandler handles endpoints for managing scenarios
type ScenarioHandler struct {
	prefix  string
	service *service.ScenarioService
	logger  *slog.Logger
}

// NewScenarioHandler creates a new instance of ScenarioHandler
func NewScenarioHandler(scenarioSvc *service.ScenarioService, logger *slog.Logger) *ScenarioHandler {
	return &ScenarioHandler{
		prefix:  "/_uni/scenarios",
		service: scenarioSvc,
		logger:  logger,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *ScenarioHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("scenario endpoint request",
		"method", r.Method,
		"path", r.URL.Path)

	path := strings.TrimPrefix(r.URL.Path, h.prefix)
	h.routeRequest(w, r, path)
}

// routeRequest routes the request to the appropriate handler method
func (h *ScenarioHandler) routeRequest(w http.ResponseWriter, r *http.Request, path string) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetRequest(w, r, path)
	case http.MethodPost:
		h.handlePostRequest(w, r, path)
	case http.MethodPut:
		h.handlePutRequest(w, r, path)
	case http.MethodDelete:
		h.handleDeleteRequest(w, r, path)
	default:
		http.NotFound(w, r)
	}
}

// handleGetRequest handles GET requests
func (h *ScenarioHandler) handleGetRequest(w http.ResponseWriter, r *http.Request, path string) {
	if path == "" {
		h.handleList(w, r)
	} else {
		uuid := strings.TrimPrefix(path, "/")
		h.handleGet(w, r, uuid)
	}
}

// handlePostRequest handles POST requests
func (h *ScenarioHandler) handlePostRequest(w http.ResponseWriter, r *http.Request, path string) {
	if path == "" {
		h.handleCreate(w, r)
	} else {
		http.NotFound(w, r)
	}
}

// handlePutRequest handles PUT requests
func (h *ScenarioHandler) handlePutRequest(w http.ResponseWriter, r *http.Request, path string) {
	if path != "" {
		uuid := strings.TrimPrefix(path, "/")
		h.handleUpdate(w, r, uuid)
	} else {
		http.NotFound(w, r)
	}
}

// handleDeleteRequest handles DELETE requests
func (h *ScenarioHandler) handleDeleteRequest(w http.ResponseWriter, r *http.Request, path string) {
	if path != "" {
		uuid := strings.TrimPrefix(path, "/")
		h.handleDelete(w, r, uuid)
	} else {
		http.NotFound(w, r)
	}
}

func (h *ScenarioHandler) handleList(w http.ResponseWriter, r *http.Request) {
	// Get all scenarios from service
	scenarios := h.service.ListScenarios(r.Context())

	// Write response
	w.Header().Set(contentTypeHeader, applicationJSON)
	err := json.NewEncoder(w).Encode(scenarios)
	if err != nil {
		h.logger.Error("failed to encode scenarios", errorLogKey, err)
	}
}

func (h *ScenarioHandler) handleGet(w http.ResponseWriter, r *http.Request, uuid string) {
	// Get the scenario from service
	scenario, err := h.service.GetScenario(r.Context(), uuid)
	if err != nil {
		h.logger.Error("failed to get scenario", errorLogKey, err, uuidLogKey, uuid)
		http.Error(w, "Scenario not found", http.StatusNotFound)
		return
	}

	// Write response
	w.Header().Set(contentTypeHeader, applicationJSON)
	err = json.NewEncoder(w).Encode(scenario)
	if err != nil {
		h.logger.Error("failed to encode scenario", errorLogKey, err)
	}
}

func (h *ScenarioHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	if !h.validateContentType(w, r.Header.Get(contentTypeHeader), "scenario creation") {
		return
	}

	scenario, err := h.parseScenarioFromBody(w, r)
	if err != nil {
		return
	}

	if err := h.service.CreateScenario(r.Context(), scenario); err != nil {
		h.handleCreateError(w, err, scenario.UUID)
		return
	}

	h.writeScenarioResponse(w, scenario, http.StatusCreated)
}

func (h *ScenarioHandler) handleUpdate(w http.ResponseWriter, r *http.Request, uuid string) {
	if !h.validateContentType(w, r.Header.Get(contentTypeHeader), "scenario update") {
		return
	}

	scenario, err := h.parseScenarioFromBody(w, r)
	if err != nil {
		return
	}

	if err := h.service.UpdateScenario(r.Context(), uuid, scenario); err != nil {
		h.handleUpdateError(w, err, uuid)
		return
	}

	h.writeScenarioResponse(w, scenario, http.StatusOK)
}

func (h *ScenarioHandler) handleDelete(w http.ResponseWriter, r *http.Request, uuid string) {
	// Delete the scenario
	if err := h.service.DeleteScenario(r.Context(), uuid); err != nil {
		h.logger.Error("failed to delete scenario", errorLogKey, err, uuidLogKey, uuid)
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

// validateContentType validates the request content type
func (h *ScenarioHandler) validateContentType(w http.ResponseWriter, contentType, operation string) bool {
	isJson := strings.HasPrefix(strings.ToLower(contentType), applicationJSON)
	if !isJson {
		h.logger.Error("invalid content type for "+operation, "content_type", contentType)
		http.Error(w, "Unsupported Media Type: Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return false
	}
	return true
}

// parseScenarioFromBody parses a scenario from the request body
func (h *ScenarioHandler) parseScenarioFromBody(w http.ResponseWriter, r *http.Request) (*model.Scenario, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read request body", errorLogKey, err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return nil, err
	}

	var scenario model.Scenario
	if err := json.Unmarshal(body, &scenario); err != nil {
		h.logger.Error("failed to unmarshal scenario", errorLogKey, err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return nil, err
	}

	return &scenario, nil
}

// handleCreateError handles errors from scenario creation
func (h *ScenarioHandler) handleCreateError(w http.ResponseWriter, err error, uuid string) {
	h.logger.Error("failed to create scenario", errorLogKey, err, uuidLogKey, uuid)
	if strings.Contains(err.Error(), "already exists") {
		http.Error(w, "Scenario already exists", http.StatusConflict)
	} else if strings.Contains(err.Error(), "invalid") {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		http.Error(w, "Failed to create scenario", http.StatusInternalServerError)
	}
}

// handleUpdateError handles errors from scenario updates
func (h *ScenarioHandler) handleUpdateError(w http.ResponseWriter, err error, uuid string) {
	h.logger.Error("failed to update scenario", errorLogKey, err, uuidLogKey, uuid)
	if strings.Contains(err.Error(), "not found") {
		http.Error(w, "Scenario not found", http.StatusNotFound)
	} else if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "mismatch") {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		http.Error(w, "Failed to update scenario", http.StatusInternalServerError)
	}
}

// writeScenarioResponse writes a scenario as JSON response
func (h *ScenarioHandler) writeScenarioResponse(w http.ResponseWriter, scenario *model.Scenario, statusCode int) {
	w.Header().Set(contentTypeHeader, applicationJSON)
	w.WriteHeader(statusCode)

	scenarioJSON, err := json.Marshal(scenario)
	if err != nil {
		h.logger.Error("failed to marshal scenario", errorLogKey, err)
		return
	}
	if _, err = w.Write(scenarioJSON); err != nil {
		h.logger.Error("failed to write scenario response", errorLogKey, err)
	}
}
