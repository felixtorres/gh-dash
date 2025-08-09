package providers

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	"github.com/charmbracelet/log"
	gh "github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"github.com/shurcooL/githubv4"

	"github.com/dlvhdr/gh-dash/v4/config"
)

type GitHubProvider struct {
	client *gh.GraphQLClient
	config ProviderConfig
}

func NewGitHubProvider(providerConfig ProviderConfig) (GitProvider, error) {
	var client *gh.GraphQLClient
	var err error

	if config.IsFeatureEnabled(config.FF_MOCK_DATA) {
		log.Debug("using mock data", "server", "https://localhost:3000")
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client, err = gh.NewGraphQLClient(gh.ClientOptions{Host: "localhost:3000", AuthToken: "fake-token"})
	} else {
		client, err = gh.DefaultGraphQLClient()
	}

	if err != nil {
		return nil, err
	}

	return &GitHubProvider{
		client: client,
		config: providerConfig,
	}, nil
}

func (p *GitHubProvider) GetType() ProviderType {
	return GitHub
}

func (p *GitHubProvider) SupportsPullRequests() bool {
	return true
}

func (p *GitHubProvider) SupportsIssues() bool {
	return true
}

func (p *GitHubProvider) GetAuthInfo() (AuthInfo, error) {
	// TODO: Implement GitHub auth info extraction
	return AuthInfo{
		Username:    "github-user",
		IsLoggedIn:  true,
		TokenSource: "GitHub CLI",
	}, nil
}

func (p *GitHubProvider) FetchPullRequests(query string, limit int, pageInfo *PageInfo) (PullRequestsResponse, error) {
	var err error
	if p.client == nil {
		if config.IsFeatureEnabled(config.FF_MOCK_DATA) {
			log.Debug("using mock data", "server", "https://localhost:3000")
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			p.client, err = gh.NewGraphQLClient(gh.ClientOptions{Host: "localhost:3000", AuthToken: "fake-token"})
		} else {
			p.client, err = gh.DefaultGraphQLClient()
		}
		if err != nil {
			return PullRequestsResponse{}, err
		}
	}

	var queryResult struct {
		Search struct {
			Nodes []struct {
				PullRequest PullRequestData `graphql:"... on PullRequest"`
			}
			IssueCount int
			PageInfo   PageInfo
		} `graphql:"search(type: ISSUE, first: $limit, after: $endCursor, query: $query)"`
	}
	var endCursor *string
	if pageInfo != nil {
		endCursor = &pageInfo.EndCursor
	}
	variables := map[string]interface{}{
		"query":     graphql.String(makePullRequestsQuery(query)),
		"limit":     graphql.Int(limit),
		"endCursor": (*graphql.String)(endCursor),
	}
	log.Debug("Fetching PRs", "query", query, "limit", limit, "endCursor", endCursor)
	err = p.client.Query("SearchPullRequests", &queryResult, variables)
	if err != nil {
		return PullRequestsResponse{}, err
	}
	log.Debug("Successfully fetched PRs", "count", queryResult.Search.IssueCount)

	prs := make([]PullRequestData, 0, len(queryResult.Search.Nodes))
	for _, node := range queryResult.Search.Nodes {
		if node.PullRequest.Repository.IsArchived {
			continue
		}
		prs = append(prs, node.PullRequest)
	}

	return PullRequestsResponse{
		Prs:        prs,
		TotalCount: queryResult.Search.IssueCount,
		PageInfo:   queryResult.Search.PageInfo,
	}, nil
}

func (p *GitHubProvider) FetchIssues(query string, limit int, pageInfo *PageInfo) (IssuesResponse, error) {
	var err error
	if p.client == nil {
		p.client, err = gh.DefaultGraphQLClient()
		if err != nil {
			return IssuesResponse{}, err
		}
	}

	var queryResult struct {
		Search struct {
			Nodes []struct {
				Issue IssueData `graphql:"... on Issue"`
			}
			IssueCount int
			PageInfo   PageInfo
		} `graphql:"search(type: ISSUE, first: $limit, after: $endCursor, query: $query)"`
	}
	var endCursor *string
	if pageInfo != nil {
		endCursor = &pageInfo.EndCursor
	}
	variables := map[string]interface{}{
		"query":     graphql.String(makeIssuesQuery(query)),
		"limit":     graphql.Int(limit),
		"endCursor": (*graphql.String)(endCursor),
	}
	log.Debug("Fetching issues", "query", query, "limit", limit, "endCursor", endCursor)
	err = p.client.Query("SearchIssues", &queryResult, variables)
	if err != nil {
		return IssuesResponse{}, err
	}
	log.Debug("Successfully fetched issues", "query", query, "count", queryResult.Search.IssueCount)

	issues := make([]IssueData, 0, len(queryResult.Search.Nodes))
	for _, node := range queryResult.Search.Nodes {
		if node.Issue.Repository.IsArchived {
			continue
		}
		issues = append(issues, node.Issue)
	}

	return IssuesResponse{
		Issues:     issues,
		TotalCount: queryResult.Search.IssueCount,
		PageInfo:   queryResult.Search.PageInfo,
	}, nil
}

func (p *GitHubProvider) FetchPullRequest(prUrl string) (PullRequestData, error) {
	var err error
	if p.client == nil {
		p.client, err = gh.DefaultGraphQLClient()
		if err != nil {
			return PullRequestData{}, err
		}
	}

	var queryResult struct {
		Resource struct {
			PullRequest PullRequestData `graphql:"... on PullRequest"`
		} `graphql:"resource(url: $url)"`
	}
	parsedUrl, err := url.Parse(prUrl)
	if err != nil {
		return PullRequestData{}, err
	}
	variables := map[string]interface{}{
		"url": githubv4.URI{URL: parsedUrl},
	}
	log.Debug("Fetching PR", "url", prUrl)
	err = p.client.Query("FetchPullRequest", &queryResult, variables)
	if err != nil {
		return PullRequestData{}, err
	}
	log.Debug("Successfully fetched PR", "url", prUrl)

	return queryResult.Resource.PullRequest, nil
}

func makePullRequestsQuery(query string) string {
	return fmt.Sprintf("is:pr %s sort:updated", query)
}

func makeIssuesQuery(query string) string {
	return fmt.Sprintf("is:issue %s sort:updated", query)
}