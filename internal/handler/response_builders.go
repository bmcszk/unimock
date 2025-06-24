package handler

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/bmcszk/unimock/pkg/config"
	"github.com/bmcszk/unimock/pkg/model"
)

// buildPOSTResponse builds response for POST requests
func (h *MockHandler) buildPOSTResponse(
	data *model.MockData,
	section *config.Section,
	sectionName string,
) (*http.Response, error) {
	// Apply response transformations
	responseData, err := h.applyResponseTransformations(data, section, sectionName)
	if err != nil {
		h.logger.Error("response transformation failed for POST", "error", err)
		return h.errorResponse(http.StatusInternalServerError, "response transformation failed"), nil
	}
	
	resp := &http.Response{
		StatusCode: http.StatusCreated,
		Header:     make(http.Header),
	}
	
	// Set Location header
	if responseData.Location != "" {
		resp.Header.Set("Location", responseData.Location)
	}
	
	// Set response body based on configuration or transformations
	if (section.Transformations != nil && section.Transformations.HasResponseTransforms()) || section.ReturnBody {
		resp.Body = io.NopCloser(bytes.NewReader(responseData.Body))
		if responseData.ContentType != "" {
			resp.Header.Set("Content-Type", responseData.ContentType)
		}
	} else {
		resp.Body = io.NopCloser(strings.NewReader(""))
	}
	
	return resp, nil
}

// buildPUTResponse builds response for PUT operations based on configuration
func (*MockHandler) buildPUTResponse(data *model.MockData, section *config.Section) *http.Response {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
	}
	
	// Set response body based on configuration
	if section.ReturnBody {
		resp.Body = io.NopCloser(bytes.NewReader(data.Body))
		if data.ContentType != "" {
			resp.Header.Set("Content-Type", data.ContentType)
		}
	} else {
		resp.Body = io.NopCloser(strings.NewReader(""))
	}
	
	if data.Location != "" {
		resp.Header.Set("Location", data.Location)
	}
	
	return resp
}

// buildDELETEResponse builds response for DELETE operations based on configuration
func (*MockHandler) buildDELETEResponse(section *config.Section) *http.Response {
	resp := &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     make(http.Header),
	}
	
	// For DELETE operations, only return body if explicitly configured
	// Most DELETE operations should return empty body regardless of transformations
	if section.ReturnBody {
		// Since DELETE doesn't have resource data to return, we return empty JSON object
		resp.Body = io.NopCloser(strings.NewReader("{}"))
		resp.Header.Set("Content-Type", "application/json")
	} else {
		resp.Body = io.NopCloser(strings.NewReader(""))
	}
	
	return resp
}

// buildSingleResourceResponse builds response for individual resource
func (*MockHandler) buildSingleResourceResponse(data *model.MockData) *http.Response {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(data.Body)),
	}
	
	if data.ContentType != "" {
		resp.Header.Set("Content-Type", data.ContentType)
	}
	
	if data.Location != "" {
		resp.Header.Set("Location", data.Location)
	}
	
	return resp
}

// buildCollectionResponse builds response for collection of resources
func (h *MockHandler) buildCollectionResponse(resources []*model.MockData) *http.Response {
	jsonItems := h.extractJSONItems(resources)
	responseBody := h.buildJSONArrayBody(jsonItems)
	
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(responseBody)),
	}
}

// extractJSONItems filters JSON resources and returns their bodies
func (*MockHandler) extractJSONItems(resources []*model.MockData) [][]byte {
	var jsonItems [][]byte
	for _, resource := range resources {
		if strings.Contains(strings.ToLower(resource.ContentType), "json") {
			jsonItems = append(jsonItems, resource.Body)
		}
	}
	return jsonItems
}

// buildJSONArrayBody creates JSON array from items
func (*MockHandler) buildJSONArrayBody(jsonItems [][]byte) []byte {
	if len(jsonItems) == 0 {
		return []byte("[]")
	}
	if len(jsonItems) == 1 {
		return buildSingleItemArray(jsonItems[0])
	}
	return buildMultiItemArray(jsonItems)
}

// buildSingleItemArray creates JSON array with single item
func buildSingleItemArray(item []byte) []byte {
	responseBody := append([]byte("["), item...)
	return append(responseBody, ']')
}

// buildMultiItemArray creates JSON array with multiple items
func buildMultiItemArray(jsonItems [][]byte) []byte {
	responseBody := []byte("[")
	for i, item := range jsonItems {
		responseBody = append(responseBody, item...)
		if i < len(jsonItems)-1 {
			responseBody = append(responseBody, ',')
		}
	}
	return append(responseBody, ']')
}

// errorResponse creates an error HTTP response
func (*MockHandler) errorResponse(statusCode int, message string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(message)),
	}
}