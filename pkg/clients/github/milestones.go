package github

import (
	"context"

	gh "github.com/google/go-github/v24/github"
	"github.com/pkg/errors"
)

// ListMilestones lists all milestones for a repository.
//
// GitHub API docs: https://developer.github.com/v3/issues/milestones/#list-milestones-for-a-repository
func (c *Client) ListMilestones(ctx context.Context) ([]*gh.Milestone, error) {
	milestones := make([]*gh.Milestone, 0)
	page := 0
	for {
		ms, resp, err := c.listMilestones(ctx, page)
		if err != nil {
			return nil, err
		}
		milestones = append(milestones, ms...)
		if resp.NextPage == 0 {
			return milestones, nil
		}
		page = resp.NextPage
	}
}

func (c *Client) listMilestones(ctx context.Context, page int) ([]*gh.Milestone, *gh.Response, error) {
	// Get all milestones
	milestones, resp, err := c.GHClient.Issues.ListMilestones(ctx, c.Owner, c.Repo, &gh.MilestoneListOptions{
		State: "all",
		ListOptions: gh.ListOptions{
			Page: page,
		},
	})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to retrieve milestones for github repository %s/%s", c.Owner, c.Repo)
	}
	return milestones, resp, nil
}
