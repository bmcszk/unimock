.PHONY: test build clean run

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary name
BINARY_NAME=unimock

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
