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

	// GetIssueFromGithubID retrieves a Jira Issue that have a 'GitHub ID' custom field matching the given github issue id.
	//
	// The returned issue may be nil if none was found.
	GetIssueFromGithubID(ghIssueID int64) (*jiralib.Issue, error)

	// UpdateIssue will update a given issue.
	//
	// JIRA API docs: https://developer.atlassian.com/cloud/jira/platform/rest/v3/?utm_source=/cloud/jira/platform/rest/&utm_medium=302#api-api-3-version-id-put
	UpdateIssue(issue *jiralib.Issue) (*jiralib.Issue, error)

	// UpdateIssueFixVersion will set the fixVersion to the given list of version ids.
	//
	// JIRA API docs: https://developer.atlassian.com/cloud/jira/platform/rest/v3/?utm_source=/cloud/jira/platform/rest/&utm_medium=302#api-api-3-version-id-put
	UpdateIssueFixVersion(issueKeyOrID string, versionsIDs []string) error

	// CreateIssue creates an issue or a sub-task from a JSON representation.
	//
	// JIRA API docs: https://docs.atlassian.com/jira/REST/latest/#api/2/issue-createIssues
	CreateIssue(issueType, summary, description, epicKey string, components []string, sprint *int, githubID int64, githubNumber int, githubLabels []string, githubStatus string) (*jiralib.Issue, error)

	// GetCustomFieldID returns a custom field ID based on its name. If not found an empty string is returned.
	GetCustomFieldID(name string) string

	// MoveToBacklog moves a list of issues identified by there issue keys to the backlog.
	// This operation is equivalent to remove future and active sprints from a given set of issues.
	//
	// JIRA API docs: https://docs.atlassian.com/jira-software/REST/7.0.4/#agile/1.0/backlog-moveIssuesToBacklog
	MoveToBacklog(issuesKeys []string) error

	// UpdateIssueEstimate estimate for the given issue key or ID
	//
	// JIRA API docs: https://developer.atlassian.com/cloud/jira/software/rest/#api-rest-agile-1-0-issue-issueIdOrKey-estimation-put
	UpdateIssueEstimate(issueKeyOrID string, estimate float32) error

	// GetIssueEstimate get the estimate for the given issue key or ID
	//
	// JIRA API docs: https://developer.atlassian.com/cloud/jira/software/rest/#api-rest-agile-1-0-issue-issueIdOrKey-estimation-get
	GetIssueEstimate(issueKeyOrID string) (float32, error)

	// AddRemoteLinkToIssue adds a link to an issue
	//
	// JIRA API docs: https://developer.atlassian.com/cloud/jira/platform/rest/v3/?utm_source=/cloud/jira/platform/rest/&utm_medium=302#api-rest-api-3-issue-issueIdOrKey-remotelink-post
	AddRemoteLinkToIssue(issueKeyOrID, globalID, title, url string) error

	// GetIssueRemoteLinks get the list of remote links for a given issue
	//
	// JIRA API docs: https://developer.atlassian.com/cloud/jira/platform/rest/v3/?utm_source=/cloud/jira/platform/rest/&utm_medium=302#api-rest-api-3-issue-issueIdOrKey-remotelink-get
	GetIssueRemoteLinks(issueKeyOrID string) ([]RemoteLink, error)

	// TransitionIssue execute transition identified by the given name to the issue
	TransitionIssue(issueKeyOrID, transitionName string) error

	// AddComment adds a new comment to issueID.
	//
	// JIRA API docs: https://docs.atlassian.com/jira/REST/latest/#api/2/issue-addComment
	AddComment(issueKeyOrID, body string) (*jiralib.Comment, error)

	// UpdateComment updates the body of a comment, identified by commentID, on the issueID.
	//
	// JIRA API docs: https://docs.atlassian.com/jira/REST/cloud/#api/2/issue/{issueIdOrKey}/comment-updateComment
	UpdateComment(issueKeyOrID, commentID, body string) (*jiralib.Comment, error)
}

// Version represents a Jira Version
type Version struct {
	jiralib.Version
	StartDate string `json:"startDate,omitempty"`
}

// dateFormat is the format used for the Last Issue Sync Update field
const issueSyncDateFormat = "2006-01-02T15:04:05.0-0700"

// RemoteLink is an issue remote link
type RemoteLink struct {
	GlobalID string           `json:"global_id,omitempty"`
	Object   RemoteLinkObject `json:"object,omitempty"`
}

// RemoteLinkObject is an issue remote link content
type RemoteLinkObject struct {
	Title string `json:"title,omitempty"`
	URL   string `json:"url,omitempty"`
}
