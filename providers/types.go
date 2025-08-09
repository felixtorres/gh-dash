package providers

import (
	"time"

	"github.com/dlvhdr/gh-dash/v4/ui/theme"
)

// Common data types that both providers use - avoiding circular imports

type Repository struct {
	Name          string
	NameWithOwner string
	IsArchived    bool
}

type Assignees struct {
	Nodes      []Assignee
	TotalCount int
}

type Assignee struct {
	Login string
}

type Comments struct {
	Nodes      []Comment
	TotalCount int
}

type Comment struct {
	Author struct {
		Login string
	}
	Body      string
	UpdatedAt time.Time
}

type Reviews struct {
	TotalCount int
	Nodes      []Review
}

type Review struct {
	Author struct {
		Login string
	}
	Body      string
	State     string
	UpdatedAt time.Time
}

type ReviewThreads struct {
	Nodes []struct {
		Id           string
		IsOutdated   bool
		OriginalLine int
		StartLine    int
		Line         int
		Path         string
		Comments     ReviewComments `graphql:"comments(first: 10)"`
	}
}

type ReviewComments struct {
	Nodes      []ReviewComment
	TotalCount int
}

type ReviewComment struct {
	Author struct {
		Login string
	}
	Body      string
	UpdatedAt time.Time
	StartLine int
	Line      int
}

type ReviewRequests struct {
	TotalCount int
	Nodes      []struct {
		AsCodeOwner bool `graphql:"asCodeOwner"`
	}
}

type ChangedFiles struct {
	TotalCount int
	Nodes      []ChangedFile
}

type ChangedFile struct {
	Additions  int
	Deletions  int
	Path       string
	ChangeType string
}

type Commits struct {
	Nodes      []CommitNode
	TotalCount int
}

type CommitNode struct {
	Commit struct {
		Deployments struct {
			Nodes []struct {
				Task        string
				Description string
			}
		}
		StatusCheckRollup struct {
			Contexts struct {
				TotalCount int
				Nodes      []struct {
					Typename      string      `graphql:"__typename"`
					CheckRun      CheckRun    `graphql:"... on CheckRun"`
					StatusContext StatusContext `graphql:"... on StatusContext"`
				}
			}
		}
	}
}

type CheckRun struct {
	Name       string
	Status     string
	Conclusion string
	CheckSuite struct {
		Creator struct {
			Login string
		}
		WorkflowRun struct {
			Workflow struct {
				Name string
			}
		}
	}
}

type StatusContext struct {
	Context string
	State   string
	Creator struct {
		Login string
	}
}

type PRLabels struct {
	Nodes []Label
}

type IssueLabels struct {
	Nodes []Label
}

type Label struct {
	Color string
	Name  string
}

type PageInfo struct {
	HasNextPage bool
	StartCursor string
	EndCursor   string
}

// Main data structures
type PullRequestData struct {
	Number int
	Title  string
	Body   string
	Author struct {
		Login string
	}
	AuthorAssociation string
	UpdatedAt         time.Time
	CreatedAt         time.Time
	Url               string
	State             string
	Mergeable         string
	ReviewDecision    string
	Additions         int
	Deletions         int
	HeadRefName       string
	BaseRefName       string
	HeadRepository    struct {
		Name string
	}
	HeadRef struct {
		Name string
	}
	Repository       Repository
	Assignees        Assignees
	Comments         Comments
	Reviews          Reviews
	ReviewThreads    ReviewThreads
	ReviewRequests   ReviewRequests
	Files            ChangedFiles
	IsDraft          bool
	Commits          Commits
	Labels           PRLabels
	MergeStateStatus string
}

type IssueData struct {
	Number int
	Title  string
	Body   string
	State  string
	Author struct {
		Login string
	}
	AuthorAssociation string
	UpdatedAt         time.Time
	CreatedAt         time.Time
	Url               string
	Repository        Repository
	Assignees         Assignees
	Comments          IssueComments
	Reactions         IssueReactions
	Labels            IssueLabels
}

type IssueComments struct {
	Nodes      []IssueComment
	TotalCount int
}

type IssueComment struct {
	Author struct {
		Login string
	}
	Body      string
	UpdatedAt time.Time
}

type IssueReactions struct {
	TotalCount int
}

type PullRequestsResponse struct {
	Prs        []PullRequestData
	TotalCount int
	PageInfo   PageInfo
}

type IssuesResponse struct {
	Issues     []IssueData
	TotalCount int
	PageInfo   PageInfo
}

// Helper methods to implement ItemData interface
func (data PullRequestData) GetAuthor(theme theme.Theme, showAuthorIcon bool) string {
	author := data.Author.Login
	if showAuthorIcon {
		// Use a simplified role icon function to avoid circular import
		author += " 👤" // Default icon
	}
	return author
}

func (data PullRequestData) GetTitle() string {
	return data.Title
}

func (data PullRequestData) GetRepoNameWithOwner() string {
	return data.Repository.NameWithOwner
}

func (data PullRequestData) GetNumber() int {
	return data.Number
}

func (data PullRequestData) GetUrl() string {
	return data.Url
}

func (data PullRequestData) GetUpdatedAt() time.Time {
	return data.UpdatedAt
}

func (data PullRequestData) GetCreatedAt() time.Time {
	return data.CreatedAt
}

func (data IssueData) GetAuthor(theme theme.Theme, showAuthorIcon bool) string {
	author := data.Author.Login
	if showAuthorIcon {
		// Use a simplified role icon function to avoid circular import
		author += " 👤" // Default icon
	}
	return author
}

func (data IssueData) GetTitle() string {
	return data.Title
}

func (data IssueData) GetRepoNameWithOwner() string {
	return data.Repository.NameWithOwner
}

func (data IssueData) GetNumber() int {
	return data.Number
}

func (data IssueData) GetUrl() string {
	return data.Url
}

func (data IssueData) GetUpdatedAt() time.Time {
	return data.UpdatedAt
}

func (data IssueData) GetCreatedAt() time.Time {
	return data.CreatedAt
}