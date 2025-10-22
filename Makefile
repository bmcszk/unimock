.PHONY: all test-all test-unit test-e2e test-e2e-up test-e2e-run test-e2e-down build clean run deps tidy vet lint check kind-start kind-stop helm-lint tilt-run tilt-stop tilt-ci k8s-setup install-lint install-gotestsum help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=gotestsum --junitfile unit-tests.xml --
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary name
BINARY_NAME=unimock

# Kubernetes parameters
K8S_CLUSTER_NAME=unimock

# Test parameters
TEST_FLAGS=-v -race -cover
TEST_E2E_FLAGS=-v -p 1 -count=1 -timeout=10m

# Health check parameters for E2E tests
HEALTH_CHECK_URL=http://localhost:28080/_uni/health
MAX_WAIT_SECONDS=10

SHELL := /bin/bash -e -o pipefail

all: build

test-all:
	@echo "Running all tests (unit and E2E)..."
	make test-unit
	make test-e2e

test-unit:
	@echo "Running unit tests..."
	$(GOTEST) $(TEST_FLAGS) ./pkg/... ./internal/...

test-e2e:
	$(MAKE) test-e2e-up
	$(MAKE) test-e2e-run
	$(MAKE) test-e2e-down

test-e2e-up:
	@echo "Starting unimock in Docker for E2E tests..."
	docker compose -f docker-compose.test.yml up -d --wait --force-recreate

test-e2e-run:
	@echo "Running E2E tests..."
	UNIMOCK_BASE_URL=http://localhost:28080 $(GOTEST) $(TEST_E2E_FLAGS) ./e2e/

test-e2e-down:
	@echo "Stopping unimock Docker containers..."
	docker compose -f docker-compose.test.yml down --remove-orphans -v || true

build:
	$(GOBUILD) -o $(BINARY_NAME) .

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

run: build
	./$(BINARY_NAME)

deps:
	$(GOMOD) download

tidy:
	$(GOMOD) tidy

vet:
	$(GOCMD) vet ./...

lint:
	golangci-lint run ./...
	golangci-lint run ./e2e

check:
	@echo "Running checks..."
	@echo "Vet..."
	make vet
	@echo "Linting..."
	make lint
	@echo "Running unit tests..."
	make test-unit
	@echo "Checks completed."

# Kubernetes and deployment targets
kind-start:
	kind create cluster --name $(K8S_CLUSTER_NAME) || echo "Cluster already exists"
	kubectl cluster-info

kind-stop:
	kind delete cluster --name $(K8S_CLUSTER_NAME)

helm-lint:
	cd helm/unimock && helm lint .

tilt-run: kind-start ## Run Tilt for local development
	cd tilt && tilt up

tilt-stop: ## Stop Tilt and clean up resources
	cd tilt && tilt down

tilt-ci: kind-start ## Run Tilt in CI mode (non-interactive)
	cd tilt && tilt ci

k8s-setup: kind-start
	helm upgrade --install $(BINARY_NAME) ./helm/unimock

# Dependencies
install-lint: ## Install golangci-lint
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6

install-gotestsum: ## Install gotestsum
	@echo "Installing gotestsum..."
	@go install gotest.tools/gotestsum@latest

help:
	@echo "Available targets:"
	@echo "  all           - Run tests and build"
	@echo "  test-all      - Run all tests (unit and E2E) with race detection and coverage"
	@echo "  test-unit     - Run unit tests (all tests not tagged 'e2e')"
	@echo "  test-e2e      - Run complete E2E test suite (Docker Compose)"
	@echo "  test-e2e-up   - Start unimock in Docker for E2E tests"
	@echo "  test-e2e-run  - Run E2E tests against running Docker container"
	@echo "  test-e2e-down - Stop unimock Docker containers"
	@echo "  build         - Build the application"
	@echo "  clean         - Remove build artifacts"
	@echo "  run           - Build and run the application"
	@echo "  deps          - Download dependencies"
	@echo "  tidy          - Tidy up dependencies"
	@echo "  vet           - Run go vet"
	@echo "  lint          - Lint the codebase"
	@echo "  kind-start    - Create a Kind Kubernetes cluster"
	@echo "  kind-stop     - Delete the Kind Kubernetes cluster"
	@echo "  helm-lint     - Lint the Helm chart"
	@echo "  tilt-run      - Start Tilt for local development"
	@echo "  tilt-stop     - Stop Tilt and clean up resources"
	@echo "  tilt-ci       - Run Tilt in CI mode (non-interactive)"
	@echo "  k8s-setup     - Deploy to Kubernetes using Helm"
	@echo "  check         - Run all checks" 
