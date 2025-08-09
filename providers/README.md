# Git Providers Support

gh-dash now supports multiple Git providers beyond GitHub, including Azure DevOps. This document explains how to configure and use different providers.

## Supported Providers

### GitHub (Default)
- Full support for all existing features
- Uses GitHub CLI for authentication
- GraphQL API for efficient data fetching

### Azure DevOps
- Pull Requests support
- Work Items (mapped to issues)
- Personal Access Token authentication
- REST API integration

## Configuration

### Automatic Detection
gh-dash will automatically detect the provider based on your git remote URL:
- `github.com` → GitHub
- `dev.azure.com` or `*.visualstudio.com` → Azure DevOps

### Manual Configuration
You can explicitly configure a provider in your config file:

```yaml
provider:
  type: azure-devops
  organization: myorg
  project: myproject
  baseUrl: https://dev.azure.com  # optional
  token: your-pat  # optional, use env var instead
```

## Authentication

### GitHub
Uses the GitHub CLI (`gh`) authentication automatically.

### Azure DevOps
Requires a Personal Access Token with the following scopes:
- Code (read) - for repositories and pull requests
- Work Items (read) - for work items/issues

Set the token via environment variable (recommended):
```bash
export AZURE_DEVOPS_TOKEN="your-personal-access-token"
# or
export ADO_PAT="your-personal-access-token"
# or
export AZURE_PAT="your-personal-access-token"
```

## Feature Mapping

| Feature | GitHub | Azure DevOps |
|---------|--------|-------------|
| Pull Requests | ✅ | ✅ |
| Issues | ✅ | ✅ (Work Items) |
| Reviews | ✅ | 🚧 (Partial) |
| Checks/CI | ✅ | 🚧 (Planned) |
| Assignees | ✅ | ✅ |
| Labels | ✅ | 🚧 (Tags) |

## Query Syntax

### GitHub
Uses GitHub's search syntax:
- `is:open author:@me`
- `is:pr review-requested:@me`
- `is:issue assignee:@me`

### Azure DevOps
Uses Azure DevOps query syntax:
- `status:active createdBy:@me`
- `status:active reviewer:@me`
- `state:Active assignedTo:@me`
- `workItemType:Bug state:Active`

## Examples

See `examples/azure-devops-config.yml` for a complete Azure DevOps configuration example.

## Troubleshooting

### Azure DevOps Authentication Issues
1. Ensure your Personal Access Token has the correct scopes
2. Check that the token is not expired
3. Verify the organization and project names are correct

### API Limits
- GitHub: Uses GraphQL for efficient querying
- Azure DevOps: Uses REST API with default limits

### Unsupported Features
Some provider-specific features may not be available across all providers. The UI will gracefully handle missing features.