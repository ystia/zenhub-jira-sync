package jira

import (
	"context"
	"fmt"
	"time"

	jiralib "github.com/andygrunwald/go-jira"
	"github.com/pkg/errors"
)

// ListSprints will return all sprints from the associated board Id.
// This only includes sprints that the user has permission to view.
//
// JIRA API docs: https://developer.atlassian.com/cloud/jira/software/rest/#api-rest-agile-1-0-board-boardId-sprint-get
func (c *Client) ListSprints(ctx context.Context) ([]jiralib.Sprint, error) {
	startAt := 0
	sprints := make([]jiralib.Sprint, 0)
	for {
		list, _, err := c.listSprints(ctx, startAt)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get Sprints for board %q", c.BoardID)
		}
		sprints = append(sprints, list.Values...)
		if list.IsLast {
			return sprints, nil
		}
		startAt = startAt + list.MaxResults
	}

}

func (c *Client) listSprints(ctx context.Context, startAt int) (*jiralib.SprintsList, *jiralib.Response, error) {
	return c.JiraClient.Board.GetAllSprintsWithOptions(c.BoardID, &jiralib.GetAllSprintsOptions{
		State: "future,active,closed",
		SearchOptions: jiralib.SearchOptions{
			StartAt: startAt,
		},
	})
}

// UpdateSprint will update a given sprint.
//
// JIRA API docs: https://developer.atlassian.com/cloud/jira/software/rest/#api-rest-agile-1-0-sprint-sprintId-put
func (c *Client) UpdateSprint(sprint *jiralib.Sprint) (*jiralib.Sprint, error) {
	req, err := c.JiraClient.NewRequest("PUT", fmt.Sprintf("/rest/agile/1.0/sprint/%d", sprint.ID), sprint)
	if err != nil {
		return sprint, errors.Wrapf(err, "failed to update sprint %q", sprint.Name)
	}
	resp, err := c.JiraClient.Do(req, sprint)
	if err != nil {
		err = jiralib.NewJiraError(resp, err)
		return nil, errors.Wrapf(err, "failed to update sprint %q", sprint.Name)
	}
	return sprint, nil
}

// CreateSprint will create a given sprint.
//
// JIRA API docs: https://developer.atlassian.com/cloud/jira/software/rest/#api-rest-agile-1-0-sprint-post
func (c *Client) CreateSprint(name string, goal string, startDate, endDate *time.Time) (*jiralib.Sprint, error) {
	var sprintRequest struct {
		Name      string     `json:"name"`
		StartDate *time.Time `json:"startDate,omitempty"`
		EndDate   *time.Time `json:"endDate,omitempty"`
		BoardID   int        `json:"originBoardId"`
		Goal      string     `json:"goal,omitempty"`
	}

	sprintRequest.Name = name
	sprintRequest.StartDate = startDate
	sprintRequest.EndDate = endDate
	sprintRequest.BoardID = c.BoardID
	sprintRequest.Goal = goal

	req, err := c.JiraClient.NewRequest("POST", "/rest/agile/1.0/sprint", &sprintRequest)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create sprint %q", name)
	}

	sprint := new(jiralib.Sprint)
	resp, err := c.JiraClient.Do(req, sprint)
	if err != nil {
		err = jiralib.NewJiraError(resp, err)
		return nil, errors.Wrapf(err, "failed to create sprint %q", sprint.Name)
	}
	return sprint, nil
}

// MoveToBacklog moves a list of issues identified by there issue keys to the backlog.
// This operation is equivalent to remove future and active sprints from a given set of issues.
//
// JIRA API docs: https://docs.atlassian.com/jira-software/REST/7.0.4/#agile/1.0/backlog-moveIssuesToBacklog
func (c *Client) MoveToBacklog(issuesKeys []string) error {
	// JIRA API support max 50 items
	issuesLimit := 50
	var batches [][]string
	for issuesLimit < len(issuesKeys) {
		issuesKeys, batches = issuesKeys[issuesLimit:], append(batches, issuesKeys[0:issuesLimit:issuesLimit])
	}
	batches = append(batches, issuesKeys)
	for _, issues := range batches {
		issueList := struct {
			Issues []string `json:"issues,omitempty"`
		}{Issues: issues}
		req, err := c.JiraClient.NewRequest("POST", "/rest/agile/1.0/backlog/issue", issueList)
		req.Header.Set("Accept", "application/json")
		if err != nil {
			return errors.Wrapf(err, "failed to move issues %v to backlog", issues)
		}
		resp, err := c.JiraClient.Do(req, nil)
		if err != nil {
			err = jiralib.NewJiraError(resp, err)
			return errors.Wrapf(err, "failed to move issues %v to backlog", issues)
		}
	}
	return nil
}
