.PHONY: test build clean run helm-lint tilt-run kind-start kind-stop k8s-setup check

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary name
BINARY_NAME=unimock

# Kubernetes parameters
K8S_CLUSTER_NAME=unimock

# Test parameters
TEST_FLAGS=-v -race -cover
TEST_E2E_FLAGS=-v -p 1 -count=1 -tags=e2e -timeout=10m

# Health check parameters for E2E tests
HEALTH_CHECK_URL=http://localhost:8080/_uni/health
MAX_WAIT_SECONDS=10

SHELL := /bin/bash -e -o pipefail

all: build

test:
	@echo "Running all tests (unit and E2E)..."
	make test-unit
	make test-e2e

test-unit:
	@echo "Running unit tests..."
	$(GOTEST) $(TEST_FLAGS) ./...

test-e2e: build
	@echo "Stopping any existing $(BINARY_NAME) process..."
	killall $(BINARY_NAME) || true
	@echo "Starting application for E2E tests..."
	./$(BINARY_NAME) > unimock_e2e.log 2>&1 & \
	APP_PID=$! ; \
	echo "Application starting with PID $$APP_PID. Logs in unimock_e2e.log" ; \
	# Ensure the application is stopped and logs are removed on exit, interrupt, or error
	trap "echo 'Stopping application (PID $$APP_PID)...'; kill $$APP_PID 2>/dev/null || true; echo 'Application stopped.' ; exit $$LAST_EXIT_CODE" EXIT INT TERM ; \
	LAST_EXIT_CODE=0; \
	( \
	    echo "Waiting for application to become healthy at $(HEALTH_CHECK_URL)..." ; \
	    COUNT=0; \
	    SUCCESS=false; \
	    while [ $$COUNT -lt $(MAX_WAIT_SECONDS) ]; do \
	        if curl -sf $(HEALTH_CHECK_URL) > /dev/null; then \
	            SUCCESS=true; \
	            break; \
	        fi; \
	        echo "Still waiting... ($$((COUNT+1))/$(MAX_WAIT_SECONDS))"; \
	        sleep 1; \
	        COUNT=$$((COUNT + 1)); \
	    done; \
	    if [ "$$SUCCESS" = "false" ]; then \
	        echo "Application failed to start and become healthy after $(MAX_WAIT_SECONDS) seconds." ; \
	        echo "--- Application Log (unimock_e2e.log) --- " ; \
	        cat unimock_e2e.log ; \
	        echo "--- End Application Log --- " ; \
	        LAST_EXIT_CODE=1; exit 1; \
	    fi; \
	    echo "Application is healthy." ; \
	    \
	    echo "Running E2E tests..." ; \
	    $(GOTEST) $(TEST_E2E_FLAGS) ./e2e/... ; \
	    LAST_EXIT_CODE=$$? ; \
	) || LAST_EXIT_CODE=$$? ; \
	exit $$LAST_EXIT_CODE

test-short:
	$(GOTEST) -short ./...

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
	golangci-lint run --build-tags=e2e

check:
	@echo "Running checks..."
	@echo "Building..."
	$(GOBUILD) ./...
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

tilt-run: kind-start
	cd tilt && tilt up

k8s-setup: kind-start
	helm upgrade --install $(BINARY_NAME) ./helm/unimock

help:
	@echo "Available targets:"
	@echo "  all        - Run tests and build"
	@echo "  test       - Run all tests (unit and E2E) with race detection and coverage"
	@echo "  test-unit  - Run unit tests (all tests not tagged 'e2e')"
	@echo "  test-e2e   - Run end-to-end tests (tests tagged 'e2e')"
	@echo "  test-short - Run only short tests"
	@echo "  build      - Build the application"
	@echo "  clean      - Remove build artifacts"
	@echo "  run        - Build and run the application"
	@echo "  deps       - Download dependencies"
	@echo "  tidy       - Tidy up dependencies"
	@echo "  vet        - Run go vet"
	@echo "  lint       - Lint the codebase"
	@echo "  kind-start - Create a Kind Kubernetes cluster"
	@echo "  kind-stop  - Delete the Kind Kubernetes cluster"
	@echo "  helm-lint  - Lint the Helm chart"
	@echo "  tilt-run   - Start Tilt for local development"
	@echo "  k8s-setup  - Deploy to Kubernetes using Helm"
	@echo "  check      - Run all checks" 
