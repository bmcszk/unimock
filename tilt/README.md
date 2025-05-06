# Tilt Configuration for Unimock

This directory contains configuration files for running Unimock in a Kubernetes development environment using [Tilt](https://tilt.dev).

## Contents

- `Tiltfile` - Main Tilt configuration file
- `tilt-local-values.yaml` - Custom Helm values for local development

## Usage

To use Tilt with Unimock, make sure you have a Kubernetes cluster running (like kind, minikube, etc.) and run:

```bash
cd tilt
tilt up
```

This will:
1. Build the Docker image from the Dockerfile
2. Deploy Unimock to your Kubernetes cluster using the Helm chart
3. Set up port forwarding so you can access the service at http://localhost:8080
4. Watch for changes in source files and automatically rebuild/redeploy

## Features

- Live reload on code changes
- Local image building
- Development-specific configuration
- Convenient UI with links to health and metrics endpoints
- Custom commands for testing

## Custom Commands

The Tilt UI provides custom commands you can run:

- `unimock-test-user` - Create a test user via the API
- `unimock-get-health` - Check the health endpoint

## Customization

You can customize the Tilt setup by:

1. Editing `tilt-local-values.yaml` to change Helm values
2. Modifying the `Tiltfile` to add new features or change behavior 
