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

all: test build

test:
	$(GOTEST) $(TEST_FLAGS) ./...

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
	golangci-lint run

check:
	@echo "Running checks..."
	@echo "Building..."
	$(GOBUILD) ./...
	@echo "Linting..."
	golangci-lint run
	@echo "Running unit tests..."
	$(GOTEST) $(TEST_FLAGS) ./...
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
	@echo "  test       - Run all tests with race detection and coverage"
	@echo "  test-short - Run only short tests"
	@echo "  build      - Build the application"
	@echo "  clean      - Remove build artifacts"
	@echo "  run        - Build and run the application"
	@echo "  deps       - Download dependencies"
	@echo "  tidy       - Tidy up dependencies"
	@echo "  vet        - Run go vet"
	@echo "  lint       - Run golangci-lint"
	@echo "  kind-start - Create a Kind Kubernetes cluster"
	@echo "  kind-stop  - Delete the Kind Kubernetes cluster"
	@echo "  helm-lint  - Lint the Helm chart"
	@echo "  tilt-run   - Start Tilt for local development"
	@echo "  k8s-setup  - Deploy to Kubernetes using Helm"
	@echo "  check      - Run all checks" 
