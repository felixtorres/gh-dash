package providers

import (
	"time"

	"github.com/dlvhdr/gh-dash/v4/ui/theme"
)

type ProviderType string

const (
	GitHub     ProviderType = "github"
	AzureDevOps ProviderType = "azure-devops"
)

type GitProvider interface {
	GetType() ProviderType
	FetchPullRequests(query string, limit int, pageInfo *PageInfo) (PullRequestsResponse, error)
	FetchIssues(query string, limit int, pageInfo *PageInfo) (IssuesResponse, error)
	FetchPullRequest(url string) (PullRequestData, error)
	
	// Provider-specific operations
	SupportsPullRequests() bool
	SupportsIssues() bool
	GetAuthInfo() (AuthInfo, error)
	
	// Command operations - return command arguments for execution
	GetDiffCommand(prNumber int, repoNameWithOwner string) ([]string, error)
	GetCheckoutCommand(prNumber int, repoNameWithOwner string) ([]string, error)
	GetMergeCommand(prNumber int, repoNameWithOwner string) ([]string, error)
	GetCloseCommand(prNumber int, repoNameWithOwner string) ([]string, error)
	GetReopenCommand(prNumber int, repoNameWithOwner string) ([]string, error)
	GetReadyCommand(prNumber int, repoNameWithOwner string) ([]string, error)
	GetUpdateCommand(prNumber int, repoNameWithOwner string) ([]string, error)
	GetWatchChecksCommand(prNumber int, repoNameWithOwner string) ([]string, error)
}

type AuthInfo struct {
	Username    string
	IsLoggedIn  bool
	TokenSource string
}

type ProviderConfig struct {
	Type         ProviderType `yaml:"type"`
	Organization string       `yaml:"organization,omitempty"`
	Project      string       `yaml:"project,omitempty"`
	BaseURL      string       `yaml:"baseUrl,omitempty"`
	Token        string       `yaml:"token,omitempty"`
}

// Common data interfaces that both providers should implement
type ItemData interface {
	GetAuthor(theme theme.Theme, showAuthorIcon bool) string
	GetTitle() string
	GetRepoNameWithOwner() string
	GetNumber() int
	GetUrl() string
	GetUpdatedAt() time.Time
	GetCreatedAt() time.Time
}

// Provider factory
func NewProvider(config ProviderConfig) (GitProvider, error) {
	switch config.Type {
	case GitHub:
		return NewGitHubProvider(config)
	case AzureDevOps:
		return NewAzureDevOpsProvider(config)
	default:
		return NewGitHubProvider(config) // Default to GitHub for backward compatibility
	}
}

// Provider detection based on remote URL
func DetectProviderFromURL(remoteURL string) ProviderType {
	if isAzureDevOpsURL(remoteURL) {
		return AzureDevOps
	}
	return GitHub // Default to GitHub
}

func isAzureDevOpsURL(url string) bool {
	return contains(url, "dev.azure.com") || 
		   contains(url, "visualstudio.com") ||
		   contains(url, ".tfs.")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (len(substr) == 0 || findIndex(s, substr) >= 0)
}

func findIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}