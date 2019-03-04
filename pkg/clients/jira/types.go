package jira

import (
	"context"
	"time"

	jiralib "github.com/andygrunwald/go-jira"
)

// API abstracts JIRA API to things needed by this project
// This is useful for mocking.
type API interface {
	// ListSprints will return all sprints from the associated board Id.
	// This only includes sprints that the user has permission to view.
	//
	// JIRA API docs: https://developer.atlassian.com/cloud/jira/software/rest/#api-rest-agile-1-0-board-boardId-sprint-get
	ListSprints(ctx context.Context) ([]jiralib.Sprint, error)
	// CreateSprint will create a given sprint.
	//
	// JIRA API docs: https://developer.atlassian.com/cloud/jira/software/rest/#api-rest-agile-1-0-sprint-post
	CreateSprint(name string, goal string, startDate, endDate *time.Time) (*jiralib.Sprint, error)
	// UpdateSprint will update a given sprint.
	//
	// JIRA API docs: https://developer.atlassian.com/cloud/jira/software/rest/#api-rest-agile-1-0-sprint-sprintId-put
	UpdateSprint(sprint *jiralib.Sprint) (*jiralib.Sprint, error)
}

// Client manages communication with the Jira API.
type Client struct {
	JiraClient *jiralib.Client
	BoardID    int
}
