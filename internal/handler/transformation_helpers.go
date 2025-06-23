package handler

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
)

// preparePostData extracts IDs and reads body for POST requests
func (h *MockHandler) preparePostData(
	ctx context.Context,
	req *http.Request,
	section *config.Section,
	sectionName string,
) (string, *model.MockData, *http.Response) {
	// Extract IDs and handle UUID generation for POST requests
	locationID, ids, err := h.extractPostResourceIDs(ctx, req, section, sectionName)
	if err != nil {
		h.logger.Error("failed to extract IDs for POST request", pathLogKey, req.URL.Path, errorLogKey, err)
		if strings.Contains(err.Error(), "failed to parse JSON body") {
			return "", nil, &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader("invalid request: failed to parse JSON body")),
			}
		}
		return "", nil, &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader("invalid request: " + err.Error())),
		}
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		h.logger.Error("failed to read request body", errorLogKey, err)
		return "", nil, &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader("failed to read request body")),
		}
	}

	initialData := &model.MockData{
		Path:        req.URL.Path,
		IDs:         ids,
		ContentType: req.Header.Get(contentTypeHeader),
		Body:        body,
	}

	return locationID, initialData, nil
}

// processPostTransformations applies transformations and creates the resource
func (h *MockHandler) processPostTransformations(
	ctx context.Context,
	_ *http.Request,
	initialData *model.MockData,
	section *config.Section,
	sectionName string,
) (*model.MockData, *http.Response) {
	// Apply request transformations
	transformedData, err := h.transformRequest(initialData, section, sectionName)
	if err != nil {
		h.logger.Error("request transformation failed", errorLogKey, err)
		return nil, h.internalServerError("request transformation failed")
	}

	// Create resource using transformed data
	err = h.service.CreateResource(ctx, transformedData.Path, transformedData.IDs, transformedData)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil, &http.Response{
				StatusCode: http.StatusConflict,
				Body:       io.NopCloser(strings.NewReader("resource already exists")),
			}
		}
		h.logger.Error("failed to create resource", errorLogKey, err)
		return nil, h.internalServerError("failed to create resource")
	}

	return transformedData, nil
}

// buildPostResponse builds the final POST response with transformations
func (h *MockHandler) buildPostResponse(
	_ context.Context,
	req *http.Request,
	transformedData *model.MockData,
	section *config.Section,
	sectionName string,
	locationID string,
) (*http.Response, error) {
	// Apply response transformations to the created data
	responseData, err := h.transformResponse(transformedData, section, sectionName)
	if err != nil {
		h.logger.Error("response transformation failed", errorLogKey, err)
		return h.internalServerError("response transformation failed"), nil
	}

	// Build response
	resp := &http.Response{
		StatusCode: http.StatusCreated,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(responseData.Body)),
	}
	
	if responseData.ContentType != "" {
		resp.Header.Set(contentTypeHeader, responseData.ContentType)
	}
	
	if responseData.Location != "" {
		resp.Header.Set("Location", responseData.Location)
	} else {
		// Set default location using the locationID from extraction
		resp.Header.Set("Location", req.URL.Path+"/"+locationID)
	}

	return resp, nil
}

// preparePutData extracts IDs and reads body for PUT requests
func (h *MockHandler) preparePutData(
	ctx context.Context,
	req *http.Request,
	section *config.Section,
	sectionName string,
) (*model.MockData, *http.Response) {
	// Extract ID from path
	ids, err := h.extractIDs(ctx, req, section, sectionName)
	if err != nil || len(ids) == 0 {
		h.logger.Warn("failed to extract ID for PUT request", pathLogKey, req.URL.Path, errorLogKey, err)
		return nil, &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader("failed to extract ID")),
		}
	}

	// Read request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		h.logger.Error("failed to read request body", errorLogKey, err)
		return nil, &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader("failed to read request body")),
		}
	}

	initialData := &model.MockData{
		Path:        req.URL.Path,
		IDs:         ids,
		ContentType: req.Header.Get(contentTypeHeader),
		Body:        body,
	}

	return initialData, nil
}

// processPutTransformations applies transformations and updates the resource
func (h *MockHandler) processPutTransformations(
	ctx context.Context,
	_ *http.Request,
	initialData *model.MockData,
	section *config.Section,
	sectionName string,
) (*model.MockData, *http.Response) {
	// Apply request transformations
	transformedData, err := h.transformRequest(initialData, section, sectionName)
	if err != nil {
		h.logger.Error("request transformation failed", errorLogKey, err)
		return nil, h.internalServerError("request transformation failed")
	}

	// Update resource
	err = h.service.UpdateResource(ctx, transformedData.Path, transformedData.IDs[0], transformedData)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("resource not found")),
			}
		}
		h.logger.Error("failed to update resource", errorLogKey, err)
		return nil, h.internalServerError("failed to update resource")
	}

	return transformedData, nil
}

// buildPutResponse builds the final PUT response with transformations
func (h *MockHandler) buildPutResponse(
	_ context.Context,
	_ *http.Request,
	transformedData *model.MockData,
	section *config.Section,
	sectionName string,
) (*http.Response, error) {
	// Apply response transformations
	responseData, err := h.transformResponse(transformedData, section, sectionName)
	if err != nil {
		h.logger.Error("response transformation failed", errorLogKey, err)
		return h.internalServerError("response transformation failed"), nil
	}

	// Build response
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(responseData.Body)),
	}
	
	if responseData.ContentType != "" {
		resp.Header.Set(contentTypeHeader, responseData.ContentType)
	}

	return resp, nil
}