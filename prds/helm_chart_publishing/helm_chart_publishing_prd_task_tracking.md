# Task Tracking: Helm Chart Publishing to GitHub Container Registry

## Analysis Summary

### Existing Workflow Structure
- **File**: `.github/workflows/docker.yml`
- **Triggers**: Tags (`v*`) and PRs
- **Registry**: GHCR (`ghcr.io`)
- **Permissions**: Already has `packages: write`
- **Authentication**: Uses `GITHUB_TOKEN`

### Helm Chart Structure
- **Chart**: `helm/unimock/Chart.yaml`
- **Current Version**: 1.0.0
- **App Version**: 1.0.0

## Design Decisions

### 1. Workflow Extension Strategy
- **Decision**: Extend existing `docker.yml` workflow
- **Rationale**: Reuse existing GHCR authentication and permissions
- **Alternative**: Create separate workflow (rejected - duplicate setup)

### 2. Versioning Strategy
- **Decision**: Update Chart.yaml version to match Git tag
- **Implementation**: Use `yq` to update version dynamically
- **Rationale**: Maintain consistency between app and chart versions

### 3. Publishing Trigger
- **Decision**: Only publish on tag pushes (not PRs)
- **Implementation**: Use `if: github.event_name != 'pull_request'`
- **Rationale**: Avoid unnecessary chart publishing on PRs

### 4. Helm Chart Registry Path
- **Decision**: Use `oci://ghcr.io/${{ github.repository_owner }}/charts/unimock`
- **Rationale**: Separate namespace from Docker images, follows OCI conventions

## Implementation Plan

### Phase 1: Workflow Enhancement
1. Add Helm installation step
2. Add chart version update step
3. Add chart packaging step
4. Add chart publishing step

### Phase 2: Testing
1. Create test branch
2. Simulate tag workflow
3. Verify chart publishing
4. Test chart installation

### Phase 3: Documentation
1. Update deployment docs
2. Add chart installation examples
3. Update README with OCI registry usage

## Test Scenarios

### BDD Test Cases

#### Scenario 1: Chart Publishing on Tag Release
```gherkin
Given a new version tag is pushed
When the GitHub Actions workflow runs
Then the Helm chart should be packaged
And the chart version should match the tag version
And the chart should be published to GHCR
```

#### Scenario 2: No Chart Publishing on PR
```gherkin
Given a pull request is opened
When the GitHub Actions workflow runs
Then the Docker image should be built
But the Helm chart should not be published
```

#### Scenario 3: Chart Installation from GHCR
```gherkin
Given a chart is published to GHCR
When a user runs helm install from the registry
Then the chart should be downloaded and installed successfully
```

## Acceptance Criteria Checklist
- [ ] Helm chart is automatically packaged on tag release
- [ ] Chart version is updated to match Git tag
- [ ] Chart is published to GHCR OCI registry
- [ ] Workflow succeeds without errors
- [ ] Chart can be installed from GHCR
- [ ] No breaking changes to existing Docker workflow
- [ ] Documentation is updated with installation instructions