# Project Structure

This document outlines the project structure, key components, and their interactions. It is intended to help the AI assistant understand the codebase and its organization.

## Top-Level Directories

- **`.github/`**: Contains GitHub Actions workflows for CI/CD and other automation.
- **`docs/`**: Contains all project documentation, including requirements, design decisions, task lists, etc.
- **`example/`**: Contains example usage of the `unimock` library/tool.
- **`helm/`**: Contains Helm charts for Kubernetes deployment.
- **`internal/`**: Contains private application and library code. It's not meant to be imported by other projects.
    - `errors/`: Project-specific error types and handling utilities.
    - `handler/`: HTTP request handlers (e.g., Chi router handlers).
    - `logger/`: Logging setup and utilities (e.g., `slog` configuration).
    - `model/`: Core data structures/domain models for internal use.
    - `router/`: HTTP router setup and configuration (e.g., Chi router initialization).
    - `service/`: Business logic layer, orchestrating actions between handlers and storage/external services.
    - `storage/`: Data persistence logic (e.g., database interactions).
- **`pkg/`**: Contains library code that's okay to be imported by external projects.
    - `client/`: Client for interacting with the `unimock` service.
    - `config/`: Configuration loading and management (e.g., using `github.com/caarlos0/env`).
    - `model/`: Public data structures/models intended for use by clients or external packages.
- **`tilt/`**: Tilt configuration for local development environments.
- **`vendor/`**: Go module dependencies.

## Key Files in Root

- **`main.go`**: The main entry point for the `unimock` application.
- **`go.mod`, `go.sum`**: Go module files.
- **`Makefile`**: Contains make targets for common development tasks (build, test, lint, etc.).
- **`Dockerfile`**: For building the application container.
- **`README.md`**: Main project overview and instructions.
- **`config.yaml`**: Example or default configuration for `unimock`.

## Service Structure (Typical Go Application)

The project generally follows a standard Go service structure:

1.  **`handler` (e.g., `internal/handler/`)**: Receives API requests, validates them, and passes them to the `service` layer.
2.  **`service` (e.g., `internal/service/`)**: Contains the core business logic, processes data, and interacts with the `storage` layer or other internal/external services.
3.  **`storage` (e.g., `internal/storage/`)**: Handles data persistence, interacting with databases or other forms of storage.
4.  **`model` (e.g., `internal/model/` or `pkg/model/`)**: Defines the data structures used throughout the application.

This structure helps in separating concerns and making the codebase more maintainable. 
