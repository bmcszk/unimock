package e2e_test

import (
	"testing"
)

func TestScenariosOnlyConfiguration_GET_ScenarioReturnsExpectedResponse(t *testing.T) {
	// given
	configContent := `
scenarios:
  - uuid: "test-user-get"
    method: "GET"
    path: "/api/users/123"
    status_code: 200
    content_type: "application/json"
    data: |
      {
        "id": "123",
        "name": "Test User",
        "email": "test@example.com"
      }

  - uuid: "test-product-post"
    method: "POST"
    path: "/api/products"
    status_code: 201
    content_type: "application/json"
    location: "/api/products/456"
    data: |
      {
        "id": "456",
        "name": "Test Product",
        "price": 99.99
      }

  - uuid: "test-error-scenario"
    method: "GET"
    path: "/api/error"
    status_code: 500
    content_type: "application/json"
    data: |
      {
        "error": "Internal server error",
        "code": "SERVER_ERROR"
      }
`

	configFile := createTempConfigFile(t, configContent)
	_, when, then := newServerParts(t, configFile)

	// when
	when.a_get_request_is_made_to("/api/users/123")

	// then
	then.the_response_is_successful()
}

func TestScenariosOnlyConfiguration_POST_ScenarioReturnsExpectedResponseWithLocationHeader(t *testing.T) {
	// given
	configContent := `
scenarios:
  - uuid: "test-user-get"
    method: "GET"
    path: "/api/users/123"
    status_code: 200
    content_type: "application/json"
    data: |
      {
        "id": "123",
        "name": "Test User",
        "email": "test@example.com"
      }

  - uuid: "test-product-post"
    method: "POST"
    path: "/api/products"
    status_code: 201
    content_type: "application/json"
    location: "/api/products/456"
    data: |
      {
        "id": "456",
        "name": "Test Product",
        "price": 99.99
      }

  - uuid: "test-error-scenario"
    method: "GET"
    path: "/api/error"
    status_code: 500
    content_type: "application/json"
    data: |
      {
        "error": "Internal server error",
        "code": "SERVER_ERROR"
      }
`

	configFile := createTempConfigFile(t, configContent)
	_, when, then := newServerParts(t, configFile)

	// when
	when.a_post_request_is_made_to("/api/products")

	// then
	then.the_post_response_is_successful()
}

func TestScenariosOnlyConfiguration_ErrorScenarioReturnsExpectedErrorResponse(t *testing.T) {
	// given
	configContent := `
scenarios:
  - uuid: "test-user-get"
    method: "GET"
    path: "/api/users/123"
    status_code: 200
    content_type: "application/json"
    data: |
      {
        "id": "123",
        "name": "Test User",
        "email": "test@example.com"
      }

  - uuid: "test-product-post"
    method: "POST"
    path: "/api/products"
    status_code: 201
    content_type: "application/json"
    location: "/api/products/456"
    data: |
      {
        "id": "456",
        "name": "Test Product",
        "price": 99.99
      }

  - uuid: "test-error-scenario"
    method: "GET"
    path: "/api/error"
    status_code: 500
    content_type: "application/json"
    data: |
      {
        "error": "Internal server error",
        "code": "SERVER_ERROR"
      }
`

	configFile := createTempConfigFile(t, configContent)
	_, when, then := newServerParts(t, configFile)

	// when
	when.a_get_request_is_made_to("/api/error")

	// then
	then.the_response_is_error()
}

func TestScenariosOnlyConfiguration_NonScenarioPathReturns404(t *testing.T) {
	// given
	configContent := `
scenarios:
  - uuid: "test-user-get"
    method: "GET"
    path: "/api/users/123"
    status_code: 200
    content_type: "application/json"
    data: |
      {
        "id": "123",
        "name": "Test User",
        "email": "test@example.com"
      }

  - uuid: "test-product-post"
    method: "POST"
    path: "/api/products"
    status_code: 201
    content_type: "application/json"
    location: "/api/products/456"
    data: |
      {
        "id": "456",
        "name": "Test Product",
        "price": 99.99
      }

  - uuid: "test-error-scenario"
    method: "GET"
    path: "/api/error"
    status_code: 500
    content_type: "application/json"
    data: |
      {
        "error": "Internal server error",
        "code": "SERVER_ERROR"
      }
`

	configFile := createTempConfigFile(t, configContent)
	_, when, then := newServerParts(t, configFile)

	// when
	when.a_get_request_is_made_to("/api/unknown")

	// then
	then.the_response_is_not_found()
}

func TestScenariosOnlyConfiguration_HealthEndpointStillWorks(t *testing.T) {
	// given
	configContent := `
scenarios:
  - uuid: "test-user-get"
    method: "GET"
    path: "/api/users/123"
    status_code: 200
    content_type: "application/json"
    data: |
      {
        "id": "123",
        "name": "Test User",
        "email": "test@example.com"
      }

  - uuid: "test-product-post"
    method: "POST"
    path: "/api/products"
    status_code: 201
    content_type: "application/json"
    location: "/api/products/456"
    data: |
      {
        "id": "456",
        "name": "Test Product",
        "price": 99.99
      }

  - uuid: "test-error-scenario"
    method: "GET"
    path: "/api/error"
    status_code: 500
    content_type: "application/json"
    data: |
      {
        "error": "Internal server error",
        "code": "SERVER_ERROR"
      }
`

	configFile := createTempConfigFile(t, configContent)
	_, when, then := newServerParts(t, configFile)

	// when
	when.a_get_request_is_made_to("/_uni/health")

	// then
	then.the_response_is_successful()
}
