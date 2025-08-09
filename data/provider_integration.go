package data

import (
	"github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/config"
	"github.com/dlvhdr/gh-dash/v4/providers"
)

var globalProviderManager *providers.ProviderManager

// InitProviders initializes the provider system
func InitProviders(cfg *config.Config, repoPath string) error {
	if globalProviderManager == nil {
		globalProviderManager = providers.NewProviderManager()
	}
	return globalProviderManager.InitializeProvider(cfg, repoPath)
}

// GetCurrentProvider returns the current provider
func GetCurrentProvider() providers.GitProvider {
	if globalProviderManager == nil {
		return nil
	}
	return globalProviderManager.GetCurrentProvider()
}

// GetProviderInfo returns information about the current provider
func GetProviderInfo() (providers.ProviderType, providers.AuthInfo, error) {
	if globalProviderManager == nil {
		return providers.GitHub, providers.AuthInfo{}, nil
	}
	return globalProviderManager.GetProviderInfo()
}

// Conversion helpers to maintain backward compatibility
func convertProviderPRToData(pr providers.PullRequestData) PullRequestData {
	return PullRequestData{
		Number:            pr.Number,
		Title:             pr.Title,
		Body:              pr.Body,
		Author:            pr.Author,
		AuthorAssociation: pr.AuthorAssociation,
		UpdatedAt:         pr.UpdatedAt,
		CreatedAt:         pr.CreatedAt,
		Url:               pr.Url,
		State:             pr.State,
		Mergeable:         pr.Mergeable,
		ReviewDecision:    pr.ReviewDecision,
		Additions:         pr.Additions,
		Deletions:         pr.Deletions,
		HeadRefName:       pr.HeadRefName,
		BaseRefName:       pr.BaseRefName,
		HeadRepository:    pr.HeadRepository,
		HeadRef:           pr.HeadRef,
		Repository: Repository{
			Name:          pr.Repository.Name,
			NameWithOwner: pr.Repository.NameWithOwner,
			IsArchived:    pr.Repository.IsArchived,
		},
		IsDraft:          pr.IsDraft,
		MergeStateStatus: MergeStateStatus(pr.MergeStateStatus),
		// Initialize complex nested fields to avoid conversion issues
		Assignees:      Assignees{},
		Comments:       Comments{},
		Reviews:        Reviews{},
		ReviewThreads:  ReviewThreads{},
		ReviewRequests: ReviewRequests{},
		Files:          ChangedFiles{},
		Commits:        Commits{},
		Labels:         PRLabels{},
	}
}

func convertProviderIssueToData(issue providers.IssueData) IssueData {
	return IssueData{
		Number:            issue.Number,
		Title:             issue.Title,
		Body:              issue.Body,
		State:             issue.State,
		Author:            issue.Author,
		AuthorAssociation: issue.AuthorAssociation,
		UpdatedAt:         issue.UpdatedAt,
		CreatedAt:         issue.CreatedAt,
		Url:               issue.Url,
		Repository: Repository{
			Name:          issue.Repository.Name,
			NameWithOwner: issue.Repository.NameWithOwner,
			IsArchived:    issue.Repository.IsArchived,
		},
		// Initialize complex nested fields to avoid conversion issues
		Assignees: Assignees{},
		Comments:  IssueComments{},
		Reactions: IssueReactions{},
		Labels:    IssueLabels{},
	}
}

// Enhanced versions of existing functions that support multiple providers
func FetchPullRequestsWithProvider(query string, limit int, pageInfo *PageInfo) (PullRequestsResponse, error) {
	provider := GetCurrentProvider()
	if provider != nil {
		log.Debug("Using provider", "type", provider.GetType(), "supportsPRs", provider.SupportsPullRequests())
		if provider.SupportsPullRequests() {
			providerResponse, err := provider.FetchPullRequests(query, limit, (*providers.PageInfo)(pageInfo))
			if err != nil {
				log.Debug("Provider fetch failed", "error", err)
				return PullRequestsResponse{}, err
			}
			log.Debug("Provider fetch successful", "count", len(providerResponse.Prs))
		
		// Convert provider response to data response
		prs := make([]PullRequestData, len(providerResponse.Prs))
		for i, pr := range providerResponse.Prs {
			prs[i] = convertProviderPRToData(pr)
		}
		
		return PullRequestsResponse{
			Prs:        prs,
			TotalCount: providerResponse.TotalCount,
			PageInfo:   PageInfo(providerResponse.PageInfo),
		}, nil
		} else {
			log.Debug("Provider doesn't support pull requests")
		}
	} else {
		log.Debug("No provider available")
	}
	
	// Fallback to original GitHub implementation
	log.Debug("Falling back to original GitHub implementation")
	return FetchPullRequests(query, limit, pageInfo)
}

func FetchIssuesWithProvider(query string, limit int, pageInfo *PageInfo) (IssuesResponse, error) {
	provider := GetCurrentProvider()
	if provider != nil && provider.SupportsIssues() {
		providerResponse, err := provider.FetchIssues(query, limit, (*providers.PageInfo)(pageInfo))
		if err != nil {
			return IssuesResponse{}, err
		}
		
		// Convert provider response to data response
		issues := make([]IssueData, len(providerResponse.Issues))
		for i, issue := range providerResponse.Issues {
			issues[i] = convertProviderIssueToData(issue)
		}
		
		return IssuesResponse{
			Issues:     issues,
			TotalCount: providerResponse.TotalCount,
			PageInfo:   PageInfo(providerResponse.PageInfo),
		}, nil
	}
	
	// Fallback to original GitHub implementation
	return FetchIssues(query, limit, pageInfo)
}

func FetchPullRequestWithProvider(url string) (PullRequestData, error) {
	provider := GetCurrentProvider()
	if provider != nil && provider.SupportsPullRequests() {
		providerPR, err := provider.FetchPullRequest(url)
		if err != nil {
			return PullRequestData{}, err
		}
		
		return convertProviderPRToData(providerPR), nil
	}
	
	// Fallback to original GitHub implementation
	return FetchPullRequest(url)
}