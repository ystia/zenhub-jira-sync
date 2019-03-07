package github

import (
	"context"

	gh "github.com/google/go-github/v24/github"
)

// API abstracts GitHub API to things needed by this project
// This is useful for mocking.
type API interface {
	// ListMilestones lists all milestones for a repository.
	//
	// GitHub API docs: https://developer.github.com/v3/issues/milestones/#list-milestones-for-a-repository
	ListMilestones(ctx context.Context) ([]*gh.Milestone, error)
	// Get fetches a repository.
	//
	// GitHub API docs: https://developer.github.com/v3/repos/#get
	GetRepository(ctx context.Context) (*gh.Repository, error)

	// GetIssue returns a single issue.
	//
	// GitHub API docs: https://developer.github.com/v3/issues/#get-a-single-issue
	GetIssue(ctx context.Context, number int) (*gh.Issue, error)

	// GetIssueFromRepoID returns a single issue from repository id.
	//
	// GitHub API docs: https://developer.github.com/v3/issues/#get-a-single-issue
	GetIssueFromRepoID(ctx context.Context, repoID int64, number int) (*gh.Issue, error)
}

// Client manages communication with the GitHub API.
type Client struct {
	GHClient *gh.Client
	Owner    string
	Repo     string
}
