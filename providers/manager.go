package providers

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/config"
	"github.com/dlvhdr/gh-dash/v4/git"
)

type ProviderManager struct {
	providers map[string]GitProvider
	current   GitProvider
}

func NewProviderManager() *ProviderManager {
	return &ProviderManager{
		providers: make(map[string]GitProvider),
	}
}

// InitializeProvider initializes the appropriate provider based on config and git remote
func (pm *ProviderManager) InitializeProvider(cfg *config.Config, repoPath string) error {
	var providerConfig ProviderConfig

	log.Debug("Initializing provider", "repoPath", repoPath)

	// If provider is explicitly configured, use that
	if cfg.Provider != nil {
		log.Debug("Using explicit provider configuration", "type", cfg.Provider.Type)
		providerConfig = ProviderConfig{
			Type:         ProviderType(cfg.Provider.Type),
			Organization: cfg.Provider.Organization,
			Project:      cfg.Provider.Project,
			BaseURL:      cfg.Provider.BaseURL,
			Token:        cfg.Provider.Token,
		}
	} else {
		// Auto-detect provider from git remote
		log.Debug("Auto-detecting provider from git remote")
		detectedConfig, err := pm.detectProviderFromGit(repoPath)
		if err != nil {
			log.Debug("Failed to detect provider from git remote", "error", err)
			// Fall back to GitHub as default
			providerConfig = ProviderConfig{
				Type: GitHub,
			}
		} else {
			log.Debug("Detected provider", "type", detectedConfig.Type, "org", detectedConfig.Organization, "project", detectedConfig.Project)
			providerConfig = *detectedConfig
		}
	}

	// Add token from environment if not set in config
	if providerConfig.Token == "" {
		providerConfig.Token = pm.getTokenFromEnvironment(providerConfig.Type)
	}

	// Initialize the provider
	provider, err := NewProvider(providerConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize %s provider: %w", providerConfig.Type, err)
	}

	key := string(providerConfig.Type)
	pm.providers[key] = provider
	pm.current = provider

	log.Debug("Initialized provider", "type", providerConfig.Type, "organization", providerConfig.Organization)
	return nil
}

// GetCurrentProvider returns the currently active provider
func (pm *ProviderManager) GetCurrentProvider() GitProvider {
	return pm.current
}

// detectProviderFromGit detects the provider type from git remote URL
func (pm *ProviderManager) detectProviderFromGit(repoPath string) (*ProviderConfig, error) {
	if repoPath == "" {
		repoPath = "."
	}

	log.Debug("Getting git remote URL", "path", repoPath)
	remoteURL, err := git.GetOriginUrl(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get git remote URL: %w", err)
	}

	log.Debug("Got remote URL", "url", remoteURL)
	config, err := GetProviderConfigFromRemote(remoteURL)
	if err != nil {
		log.Debug("Failed to parse remote URL", "url", remoteURL, "error", err)
		return nil, err
	}

	log.Debug("Parsed provider config", "type", config.Type, "org", config.Organization, "project", config.Project)
	return config, nil
}

// getTokenFromEnvironment gets authentication token from environment variables
func (pm *ProviderManager) getTokenFromEnvironment(providerType ProviderType) string {
	switch providerType {
	case GitHub:
		// GitHub CLI handles authentication automatically
		return ""
	case AzureDevOps:
		// Look for Azure DevOps Personal Access Token
		log.Debug("Looking for Azure DevOps token in environment variables")
		if token := os.Getenv("AZURE_DEVOPS_TOKEN"); token != "" {
			log.Debug("Found AZURE_DEVOPS_TOKEN")
			return token
		}
		if token := os.Getenv("ADO_PAT"); token != "" {
			log.Debug("Found ADO_PAT")
			return token
		}
		if token := os.Getenv("AZURE_PAT"); token != "" {
			log.Debug("Found AZURE_PAT")
			return token
		}
		log.Debug("No Azure DevOps token found in environment variables")
	}
	return ""
}

// GetProviderInfo returns information about the current provider
func (pm *ProviderManager) GetProviderInfo() (ProviderType, AuthInfo, error) {
	if pm.current == nil {
		return "", AuthInfo{}, fmt.Errorf("no provider initialized")
	}

	authInfo, err := pm.current.GetAuthInfo()
	return pm.current.GetType(), authInfo, err
}