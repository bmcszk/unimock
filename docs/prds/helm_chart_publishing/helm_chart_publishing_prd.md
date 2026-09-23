# PRD: Helm Chart Publishing to GitHub Container Registry

## Overview
Enable automatic publishing of Unimock Helm charts to GitHub Container Registry (GHCR) alongside existing Docker image publishing.

## Problem Statement
Currently, Unimock publishes Docker images to GHCR but not the Helm chart. Users must manually package and install the chart from the repository, creating friction in the deployment process.

## Success Criteria
- Helm charts are automatically published to GHCR on releases
- Charts are versioned consistently with application releases
- Users can install directly from GHCR: `helm install unimock oci://ghcr.io/username/unimock`
- Integration with existing CI/CD pipeline

## Requirements

### Functional Requirements
1. **Automatic Publishing**: Helm charts published on every release/tag
2. **Version Consistency**: Chart version matches application version
3. **OCI Registry Support**: Use GHCR as OCI-compatible Helm registry
4. **Authentication**: Use existing GitHub token for publishing

### Technical Requirements
1. **Extend Existing Workflow**: Add to current `.github/workflows/release.yml`
2. **Chart Packaging**: Use `helm package` to create chart archive
3. **OCI Push**: Use `helm push` with OCI protocol
4. **Error Handling**: Fail gracefully with clear error messages
5. **Build Dependencies**: Run after successful Docker image build

### Non-Functional Requirements
1. **Performance**: Minimal impact on existing workflow duration
2. **Reliability**: Consistent publishing across all releases
3. **Security**: Use existing GHCR permissions and tokens

## Implementation Approach
1. Extend existing GitHub Actions workflow
2. Add Helm chart packaging and publishing steps
3. Use conditional logic to only publish on releases/tags
4. Maintain backward compatibility with existing workflow

## Acceptance Criteria
- [ ] Helm chart is packaged and pushed to GHCR on release
- [ ] Chart version matches Git tag version
- [ ] Workflow passes without errors
- [ ] Chart can be installed from GHCR
- [ ] No breaking changes to existing Docker publishing

## Timeline
- Design and Implementation: 1 day
- Testing and Validation: 1 day
- Documentation Update: 1 day

## Dependencies
- Existing GitHub Actions workflow
- Helm CLI availability in GitHub Actions runner
- GHCR write permissions via GitHub token