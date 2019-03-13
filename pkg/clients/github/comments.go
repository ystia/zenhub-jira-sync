package github

import (
	"context"

	"github.com/pkg/errors"

	gh "github.com/google/go-github/v24/github"
)

// GetIssueComments gets all comments on the specified issue. Specifying an issue
// number of 0 will return all comments on all issues for the repository.
//
// GitHub API docs: https://developer.github.com/v3/issues/comments/#list-comments-on-an-issue
func (c *Client) GetIssueComments(ctx context.Context, issueNumber int) ([]*gh.IssueComment, error) {
	var comments []*gh.IssueComment
	var page int
	for {
		com, resp, err := c.listIssueComments(ctx, issueNumber, page)
		if err != nil {
			return nil, err
		}
		comments = append(comments, com...)
		if resp.NextPage == 0 {
			return comments, nil
		}
		page = resp.NextPage
	}

	return nil, nil

}

func (c *Client) listIssueComments(ctx context.Context, issueNumber, page int) ([]*gh.IssueComment, *gh.Response, error) {
	comments, resp, err := c.GHClient.Issues.ListComments(ctx, c.Owner, c.Repo, issueNumber, &gh.IssueListCommentsOptions{
		ListOptions: gh.ListOptions{
			Page: page,
		},
	})
	return comments, resp, errors.Wrapf(err, "failed to list comments for issue #%d", issueNumber)
}
