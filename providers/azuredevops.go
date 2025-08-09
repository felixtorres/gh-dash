package providers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/ui/theme"
)

type AzureDevOpsProvider struct {
	client       *http.Client
	config       ProviderConfig
	organization string
	project      string
	baseURL      string
	token        string
}

type AzurePullRequest struct {
	PullRequestId int    `json:"pullRequestId"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Status        string `json:"status"`
	CreatedBy     struct {
		DisplayName string `json:"displayName"`
		UniqueName  string `json:"uniqueName"`
	} `json:"createdBy"`
	CreationDate   time.Time `json:"creationDate"`
	SourceRefName  string    `json:"sourceRefName"`
	TargetRefName  string    `json:"targetRefName"`
	Repository     struct {
		Name    string `json:"name"`
		WebURL  string `json:"webUrl"`
		Project struct {
			Name string `json:"name"`
		} `json:"project"`
	} `json:"repository"`
	Url string `json:"url"`
}

type AzureWorkItem struct {
	Id     int `json:"id"`
	Fields struct {
		Title       string    `json:"System.Title"`
		State       string    `json:"System.State"`
		CreatedBy   string    `json:"System.CreatedBy"`
		CreatedDate time.Time `json:"System.CreatedDate"`
		ChangedDate time.Time `json:"System.ChangedDate"`
		Description string    `json:"System.Description"`
	} `json:"fields"`
	Url string `json:"url"`
}

type AzurePullRequestsResponse struct {
	Value []AzurePullRequest `json:"value"`
	Count int                `json:"count"`
}

type AzureWorkItemsResponse struct {
	Value []AzureWorkItem `json:"value"`
	Count int             `json:"count"`
}

func NewAzureDevOpsProvider(config ProviderConfig) (GitProvider, error) {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://dev.azure.com"
	}

	log.Debug("Creating Azure DevOps provider", "org", config.Organization, "project", config.Project, "baseURL", baseURL, "hasToken", config.Token != "")

	return &AzureDevOpsProvider{
		client:       &http.Client{Timeout: 30 * time.Second},
		config:       config,
		organization: config.Organization,
		project:      config.Project,
		baseURL:      baseURL,
		token:        config.Token,
	}, nil
}

func (p *AzureDevOpsProvider) GetType() ProviderType {
	return AzureDevOps
}

func (p *AzureDevOpsProvider) SupportsPullRequests() bool {
	return true
}

func (p *AzureDevOpsProvider) SupportsIssues() bool {
	return true // Azure DevOps work items map to issues
}

func (p *AzureDevOpsProvider) GetAuthInfo() (AuthInfo, error) {
	return AuthInfo{
		Username:    "azure-user",
		IsLoggedIn:  p.token != "",
		TokenSource: "Personal Access Token",
	}, nil
}

func (p *AzureDevOpsProvider) FetchPullRequests(query string, limit int, pageInfo *PageInfo) (PullRequestsResponse, error) {
	log.Debug("Azure DevOps FetchPullRequests called", "query", query, "limit", limit, "org", p.organization, "project", p.project)

	// Check if we have required configuration
	if p.organization == "" || p.project == "" {
		return PullRequestsResponse{}, fmt.Errorf("Azure DevOps organization and project are required")
	}

	// Check if we have authentication
	if p.token == "" {
		return PullRequestsResponse{}, fmt.Errorf("Azure DevOps Personal Access Token is required. Set AZURE_DEVOPS_TOKEN, ADO_PAT, or AZURE_PAT environment variable")
	}

	apiURL := fmt.Sprintf("%s/%s/%s/_apis/git/pullrequests?api-version=7.1&$top=%d", 
		p.baseURL, p.organization, p.project, limit)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return PullRequestsResponse{}, err
	}

	p.setAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	log.Debug("Fetching Azure DevOps PRs", "url", apiURL, "limit", limit)
	
	resp, err := p.client.Do(req)
	if err != nil {
		return PullRequestsResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return PullRequestsResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return PullRequestsResponse{}, fmt.Errorf("Azure DevOps API error: %s", string(body))
	}

	var azureResp AzurePullRequestsResponse
	if err := json.Unmarshal(body, &azureResp); err != nil {
		return PullRequestsResponse{}, err
	}

	log.Debug("Successfully fetched Azure DevOps PRs", "count", azureResp.Count)

	prs := make([]PullRequestData, 0, len(azureResp.Value))
	for _, azurePR := range azureResp.Value {
		// Convert Azure DevOps PR to common format
		pr := p.convertAzurePRToData(azurePR)
		prs = append(prs, pr)
	}

	return PullRequestsResponse{
		Prs:        prs,
		TotalCount: azureResp.Count,
		PageInfo:   PageInfo{}, // TODO: Implement pagination
	}, nil
}

func (p *AzureDevOpsProvider) FetchIssues(query string, limit int, pageInfo *PageInfo) (IssuesResponse, error) {
	// Use Work Items API for issues
	wiql := fmt.Sprintf("SELECT [System.Id], [System.Title], [System.State], [System.CreatedBy], [System.CreatedDate], [System.ChangedDate] FROM workitems WHERE [System.TeamProject] = '%s'", p.project)
	
	apiURL := fmt.Sprintf("%s/%s/%s/_apis/wit/wiql?api-version=7.1", 
		p.baseURL, p.organization, p.project)

	wiqlQuery := map[string]string{"query": wiql}
	jsonData, err := json.Marshal(wiqlQuery)
	if err != nil {
		return IssuesResponse{}, err
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return IssuesResponse{}, err
	}

	p.setAuthHeader(req)
	req.Header.Set("Content-Type", "application/json")

	log.Debug("Fetching Azure DevOps work items", "url", apiURL, "limit", limit)
	
	resp, err := p.client.Do(req)
	if err != nil {
		return IssuesResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return IssuesResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return IssuesResponse{}, fmt.Errorf("Azure DevOps API error: %s", string(body))
	}

	// TODO: Parse WIQL response and fetch actual work item details
	// This is a simplified implementation
	return IssuesResponse{
		Issues:     []IssueData{},
		TotalCount: 0,
		PageInfo:   PageInfo{},
	}, nil
}

func (p *AzureDevOpsProvider) FetchPullRequest(prUrl string) (PullRequestData, error) {
	// Extract PR ID from URL and fetch individual PR
	// This is a simplified implementation
	return PullRequestData{}, fmt.Errorf("FetchPullRequest not implemented for Azure DevOps")
}

func (p *AzureDevOpsProvider) setAuthHeader(req *http.Request) {
	if p.token != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(":" + p.token))
		req.Header.Set("Authorization", "Basic "+auth)
	}
}

func (p *AzureDevOpsProvider) convertAzurePRToData(azurePR AzurePullRequest) PullRequestData {
	// Map Azure DevOps states to GitHub-like states
	state := "open"
	if azurePR.Status == "completed" {
		state = "merged"
	} else if azurePR.Status == "abandoned" {
		state = "closed"
	}

	return PullRequestData{
		Number: azurePR.PullRequestId,
		Title:  azurePR.Title,
		Body:   azurePR.Description,
		Author: struct {
			Login string
		}{Login: azurePR.CreatedBy.DisplayName},
		AuthorAssociation: "MEMBER", // Default assumption
		UpdatedAt:         azurePR.CreationDate, // Azure doesn't separate created/updated easily
		CreatedAt:         azurePR.CreationDate,
		Url:               azurePR.Url,
		State:             state,
		Mergeable:         "UNKNOWN",
		ReviewDecision:    "",
		Additions:         0, // Would need additional API call
		Deletions:         0, // Would need additional API call
		HeadRefName:       strings.TrimPrefix(azurePR.SourceRefName, "refs/heads/"),
		BaseRefName:       strings.TrimPrefix(azurePR.TargetRefName, "refs/heads/"),
		HeadRepository: struct {
			Name string
		}{Name: azurePR.Repository.Name},
		HeadRef: struct {
			Name string
		}{Name: strings.TrimPrefix(azurePR.SourceRefName, "refs/heads/")},
		Repository: Repository{
			Name:          azurePR.Repository.Name,
			NameWithOwner: fmt.Sprintf("%s/%s", azurePR.Repository.Project.Name, azurePR.Repository.Name),
			IsArchived:    false,
		},
		// Initialize other fields with default values
		Assignees:        Assignees{},
		Comments:         Comments{},
		Reviews:          Reviews{},
		ReviewThreads:    ReviewThreads{},
		ReviewRequests:   ReviewRequests{},
		Files:            ChangedFiles{},
		IsDraft:          false,
		Commits:          Commits{},
		Labels:           PRLabels{},
		MergeStateStatus: "",
	}
}

// Helper type to implement ItemData interface for Azure DevOps items
type AzureDevOpsItemData struct {
	title         string
	author        string
	repoName      string
	number        int
	url           string
	updatedAt     time.Time
	createdAt     time.Time
	authorAssoc   string
}

func (d AzureDevOpsItemData) GetAuthor(theme theme.Theme, showAuthorIcon bool) string {
	author := d.author
	if showAuthorIcon {
		// Use a simplified role icon since we don't want circular imports
		author += " 👤" // Default user icon
	}
	return author
}

func (d AzureDevOpsItemData) GetTitle() string {
	return d.title
}

func (d AzureDevOpsItemData) GetRepoNameWithOwner() string {
	return d.repoName
}

func (d AzureDevOpsItemData) GetNumber() int {
	return d.number
}

func (d AzureDevOpsItemData) GetUrl() string {
	return d.url
}

func (d AzureDevOpsItemData) GetUpdatedAt() time.Time {
	return d.updatedAt
}

func (d AzureDevOpsItemData) GetCreatedAt() time.Time {
	return d.createdAt
}