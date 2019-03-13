package jira

import (
	jiralib "github.com/andygrunwald/go-jira"
	"github.com/pkg/errors"
)

// AddComment adds a new comment to issueID.
//
// JIRA API docs: https://docs.atlassian.com/jira/REST/latest/#api/2/issue-addComment
func (c *Client) AddComment(issueKeyOrID, body string) (*jiralib.Comment, error) {

	comment, _, err := c.JiraClient.Issue.AddComment(issueKeyOrID, &jiralib.Comment{
		Body: body,
	})
	return comment, errors.Wrapf(err, "failed to create comment for issue %q", issueKeyOrID)

}

// UpdateComment updates the body of a comment, identified by commentID, on the issueID.
//
// JIRA API docs: https://docs.atlassian.com/jira/REST/cloud/#api/2/issue/{issueIdOrKey}/comment-updateComment
func (c *Client) UpdateComment(issueKeyOrID, commentID, body string) (*jiralib.Comment, error) {

	comment, _, err := c.JiraClient.Issue.UpdateComment(issueKeyOrID, &jiralib.Comment{
		ID:   commentID,
		Body: body,
	})
	return comment, errors.Wrapf(err, "failed to update comment for issue %q", issueKeyOrID)
}
