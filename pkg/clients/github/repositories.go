package github

import (
	"context"

	"github.com/pkg/errors"

	gh "github.com/google/go-github/v24/github"
)

// GetRepository fetches a repository.
//
// GitHub API docs: https://developer.github.com/v3/repos/#get
func (c *Client) GetRepository(ctx context.Context) (*gh.Repository, error) {
	repo, _, err := c.GHClient.Repositories.Get(ctx, c.Owner, c.Repo)
	return repo, errors.Wrapf(err, "failed to get %s/%s github repository", c.Owner, c.Repo)
}
