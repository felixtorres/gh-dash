# Azure DevOps Support Implementation

## Overview

This implementation extends gh-dash to support Azure DevOps in addition to GitHub, providing a multi-provider architecture that allows users to view pull requests and work items from Azure DevOps repositories.

## Architecture Changes

### 1. Provider Abstraction Layer (`providers/`)

- **`provider.go`**: Core interfaces and provider factory
- **`types.go`**: Common data structures shared across providers
- **`github.go`**: GitHub provider implementation
- **`azuredevops.go`**: Azure DevOps provider implementation
- **`detection.go`**: Auto-detection of provider from git remote URLs
- **`manager.go`**: Provider management and initialization

### 2. Enhanced Configuration

- Added `provider` section to config files
- Support for Azure DevOps-specific settings (organization, project, token)
- Automatic provider detection from git remote URLs

### 3. Data Layer Integration

- **`provider_integration.go`**: Bridges provider system with existing data layer
- Maintains backward compatibility with existing GitHub-only functionality
- Provides enhanced functions that use the provider system

## Features Implemented

### Azure DevOps Support
- ✅ Pull Requests fetching via REST API
- ✅ Work Items as Issues (basic structure)
- ✅ Personal Access Token authentication
- ✅ Organization and Project configuration
- ✅ Auto-detection from git remotes

### Provider System
- ✅ Multi-provider architecture
- ✅ Provider auto-detection
- ✅ Fallback to GitHub for backward compatibility
- ✅ Configuration-based provider selection

### URL Pattern Support
- GitHub: `https://github.com/owner/repo`
- Azure DevOps: `https://dev.azure.com/org/project/_git/repo`
- Visual Studio: `https://org.visualstudio.com/project/_git/repo`
- On-premises TFS: `https://server/tfs/collection/project/_git/repo`

## Configuration Examples

### Auto-detection (Recommended)
```yaml
# No provider config needed - will auto-detect from git remote
prSections:
  - title: My Pull Requests
    filters: "is:open author:@me"
```

### Explicit Azure DevOps Configuration
```yaml
provider:
  type: azure-devops
  organization: myorg
  project: myproject
  # token: set via environment variable

prSections:
  - title: My Pull Requests
    filters: "createdBy:@me status:active"
```

### Environment Variables
```bash
export AZURE_DEVOPS_TOKEN="your-personal-access-token"
# or
export ADO_PAT="your-personal-access-token"
# or  
export AZURE_PAT="your-personal-access-token"
```

## Usage Examples

1. **GitHub Repository (existing behavior)**:
   ```bash
   cd my-github-repo
   gh dash
   # Auto-detects GitHub, uses existing functionality
   ```

2. **Azure DevOps Repository**:
   ```bash
   cd my-azure-repo  
   gh dash
   # Auto-detects Azure DevOps, uses new provider
   ```

3. **Explicit Configuration**:
   ```bash
   gh dash -c azure-devops-config.yml
   # Uses explicit provider configuration
   ```

## Implementation Details

### Provider Detection Flow
1. Check if provider is explicitly configured
2. If not, analyze git remote URL to detect provider type
3. Extract organization, project, and repository information
4. Initialize appropriate provider with detected configuration
5. Fallback to GitHub if detection fails

### Data Flow
1. UI requests data through existing data layer functions
2. Data layer checks if provider system is initialized
3. If yes, routes request to appropriate provider
4. Provider fetches data via its API (GraphQL for GitHub, REST for Azure DevOps)
5. Response is converted to common format
6. Data is returned to UI in expected format

### Backward Compatibility
- Existing GitHub-only configurations continue to work unchanged
- No breaking changes to existing APIs
- GitHub remains the default provider
- All existing features continue to function

## Testing

The implementation successfully:
- ✅ Compiles without errors
- ✅ Maintains existing GitHub functionality
- ✅ Supports basic Azure DevOps operations
- ✅ Handles provider auto-detection
- ✅ Processes configuration correctly

## Future Enhancements

### Azure DevOps Improvements
- Full Work Items API integration
- Azure DevOps-specific query syntax
- Pull Request reviews and comments
- Build/Pipeline status integration
- Enhanced authentication methods

### Additional Providers
- GitLab support
- Bitbucket support  
- Generic Git provider interface

### Advanced Features
- Multi-provider dashboards
- Cross-provider search
- Provider-specific themes and layouts
- Advanced caching strategies

## Files Modified/Created

### New Files
- `providers/provider.go` - Core provider interfaces
- `providers/types.go` - Common data structures
- `providers/github.go` - GitHub provider implementation
- `providers/azuredevops.go` - Azure DevOps provider implementation
- `providers/detection.go` - Provider detection logic
- `providers/manager.go` - Provider management
- `data/provider_integration.go` - Integration layer
- `examples/azure-devops-config.yml` - Example configuration
- `providers/README.md` - Provider documentation

### Modified Files
- `config/parser.go` - Added provider configuration support
- `git/git.go` - Enhanced repository name parsing for multiple providers

This implementation provides a solid foundation for multi-provider support while maintaining full backward compatibility with existing GitHub-based workflows.