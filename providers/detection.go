package providers

import (
	"regexp"
	"strings"
)

// ParseRemoteURL extracts provider information from a git remote URL
type RemoteInfo struct {
	Provider     ProviderType
	Organization string
	Project      string
	Repository   string
	BaseURL      string
}

// ParseGitRemoteURL parses a git remote URL and returns provider information
func ParseGitRemoteURL(remoteURL string) (*RemoteInfo, error) {
	// Clean the URL
	url := strings.TrimSpace(remoteURL)
	
	// Handle SSH URLs by converting them to HTTPS format for parsing
	if strings.HasPrefix(url, "git@") {
		url = convertSSHToHTTPS(url)
	}
	
	// Detect Azure DevOps
	if azureInfo := parseAzureDevOpsURL(url); azureInfo != nil {
		return azureInfo, nil
	}
	
	// Default to GitHub
	return parseGitHubURL(url), nil
}

func convertSSHToHTTPS(sshURL string) string {
	// Convert git@github.com:owner/repo.git to https://github.com/owner/repo.git
	// Convert git@ssh.dev.azure.com:v3/org/project/repo to https://dev.azure.com/org/project/_git/repo
	
	if strings.Contains(sshURL, "github.com") {
		re := regexp.MustCompile(`git@github\.com:([^/]+)/(.+)\.git`)
		matches := re.FindStringSubmatch(sshURL)
		if len(matches) == 3 {
			return "https://github.com/" + matches[1] + "/" + matches[2] + ".git"
		}
	}
	
	if strings.Contains(sshURL, "dev.azure.com") || strings.Contains(sshURL, "ssh.dev.azure.com") {
		re := regexp.MustCompile(`git@ssh\.dev\.azure\.com:v3/([^/]+)/([^/]+)/(.+)`)
		matches := re.FindStringSubmatch(sshURL)
		if len(matches) == 4 {
			return "https://dev.azure.com/" + matches[1] + "/" + matches[2] + "/_git/" + matches[3]
		}
	}
	
	return sshURL
}

func parseAzureDevOpsURL(url string) *RemoteInfo {
	// Azure DevOps patterns:
	// https://dev.azure.com/{organization}/{project}/_git/{repository}
	// https://{organization}.visualstudio.com/{project}/_git/{repository}
	// https://{server}/{tfs}/{collection}/{project}/_git/{repository}
	
	patterns := []string{
		`https://dev\.azure\.com/([^/]+)/([^/]+)/_git/(.+?)(?:\.git)?/?$`,
		`https://([^.]+)\.visualstudio\.com/([^/]+)/_git/(.+?)(?:\.git)?/?$`,
		`https://([^/]+)/tfs/([^/]+)/([^/]+)/_git/(.+?)(?:\.git)?/?$`,
	}
	
	for i, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(url)
		
		switch i {
		case 0: // dev.azure.com
			if len(matches) == 4 {
				return &RemoteInfo{
					Provider:     AzureDevOps,
					Organization: matches[1],
					Project:      matches[2],
					Repository:   matches[3],
					BaseURL:      "https://dev.azure.com",
				}
			}
		case 1: // visualstudio.com
			if len(matches) == 4 {
				return &RemoteInfo{
					Provider:     AzureDevOps,
					Organization: matches[1],
					Project:      matches[2],
					Repository:   matches[3],
					BaseURL:      "https://" + matches[1] + ".visualstudio.com",
				}
			}
		case 2: // on-premises TFS
			if len(matches) == 5 {
				return &RemoteInfo{
					Provider:     AzureDevOps,
					Organization: matches[2], // collection
					Project:      matches[3],
					Repository:   matches[4],
					BaseURL:      "https://" + matches[1],
				}
			}
		}
	}
	
	return nil
}

func parseGitHubURL(url string) *RemoteInfo {
	// GitHub patterns:
	// https://github.com/{owner}/{repository}
	// git@github.com:{owner}/{repository}.git
	
	re := regexp.MustCompile(`https://github\.com/([^/]+)/(.+?)(?:\.git)?/?$`)
	matches := re.FindStringSubmatch(url)
	
	if len(matches) == 3 {
		return &RemoteInfo{
			Provider:     GitHub,
			Organization: matches[1],
			Project:      "", // GitHub doesn't have a separate project concept
			Repository:   matches[2],
			BaseURL:      "https://github.com",
		}
	}
	
	// Fallback - assume GitHub format
	return &RemoteInfo{
		Provider:     GitHub,
		Organization: "unknown",
		Project:      "",
		Repository:   "unknown",
		BaseURL:      "https://github.com",
	}
}

// GetProviderConfigFromRemote creates a provider config from remote URL
func GetProviderConfigFromRemote(remoteURL string) (*ProviderConfig, error) {
	info, err := ParseGitRemoteURL(remoteURL)
	if err != nil {
		return nil, err
	}
	
	config := &ProviderConfig{
		Type:         info.Provider,
		Organization: info.Organization,
		Project:      info.Project,
		BaseURL:      info.BaseURL,
	}
	
	return config, nil
}