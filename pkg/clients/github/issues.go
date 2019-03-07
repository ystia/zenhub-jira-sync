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
