package github

import (
	"context"

	"github.com/pkg/errors"

	gh "github.com/google/go-github/v24/github"
)

// GetIssue returns a single issue.
//
// GitHub API docs: https://developer.github.com/v3/issues/#get-a-single-issue
func (c *Client) GetIssue(ctx context.Context, number int) (*gh.Issue, error) {
	issue, _, err := c.GHClient.Issues.Get(ctx, c.Owner, c.Repo, number)
	return issue, errors.Wrapf(err, "failed to get issue %s/%s#%d", c.Owner, c.Repo, number)
}

// GetIssueFromRepoID returns a single issue from repository id.
//
// GitHub API docs: https://developer.github.com/v3/issues/#get-a-single-issue
func (c *Client) GetIssueFromRepoID(ctx context.Context, repoID int64, number int) (*gh.Issue, error) {
	repo, _, err := c.GHClient.Repositories.GetByID(ctx, repoID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get info for repository with id %d", repoID)
	}
	issue, _, err := c.GHClient.Issues.Get(ctx, repo.GetOwner().GetLogin(), repo.GetName(), number)
	return issue, errors.Wrapf(err, "failed to get issue %s/%s#%d", repo.GetOwner().GetLogin(), repo.GetName(), number)
}

// ListIssues lists the issues for the specified repository.
//
// GitHub API docs: https://developer.github.com/v3/issues/#list-issues-for-a-repository
func (c *Client) ListIssues(ctx context.Context, opts *gh.IssueListByRepoOptions) ([]*gh.Issue, error) {
	if opts == nil {
		opts = &gh.IssueListByRepoOptions{}
	}
	issues := make([]*gh.Issue, 0)
	opts.Page = 0
	for {
		i, resp, err := c.listIssues(ctx, opts)
		if err != nil {
			return nil, err
		}
		issues = append(issues, i...)
		if resp.NextPage == 0 {
			return issues, nil
		}
		opts.Page = resp.NextPage
	}
}

func (c *Client) listIssues(ctx context.Context, opts *gh.IssueListByRepoOptions) ([]*gh.Issue, *gh.Response, error) {
	issues, resp, err := c.GHClient.Issues.ListByRepo(ctx, c.Owner, c.Repo, opts)
	return issues, resp, errors.Wrapf(err, "failed to list issues for repository %s/%s", c.Owner, c.Repo)
}
