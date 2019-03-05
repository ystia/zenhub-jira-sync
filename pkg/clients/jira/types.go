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
	// GetProjectVersions returns all versions of the associated Jira project
	GetProjectVersions() ([]*Version, error)
	// Create creates a version in JIRA.
	//
	// JIRA API docs: https://developer.atlassian.com/cloud/jira/platform/rest/v3/?utm_source=%2Fcloud%2Fjira%2Fplatform%2Frest%2F&utm_medium=302#api-api-3-version-post
	CreateVersion(name, description string, projectID int, released, archived bool, startDate, dueDate, releaseDate *time.Time) (*Version, error)
	// UpdateVersion will update a given version.
	//
	// JIRA API docs: https://developer.atlassian.com/cloud/jira/platform/rest/v3/?utm_source=/cloud/jira/platform/rest/&utm_medium=302#api-api-3-version-id-put
	UpdateVersion(version *Version) (*Version, error)

	// GetProjectID
	GetProjectID() (int, error)
}

// Client manages communication with the Jira API.
type Client struct {
	JiraClient *jiralib.Client
	ProjectID  string
	BoardID    int
}

// Version represents a Jira Version
type Version struct {
	jiralib.Version
	StartDate string `json:"startDate,omitempty"`
}
