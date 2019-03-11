package jira

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/pkg/errors"

	jiralib "github.com/andygrunwald/go-jira"
)

// GetIssueFromGithubID retrieves a Jira Issue that have a 'GitHub ID' custom field matching the given github issue id.
//
// The returned issue may be nil if none was found.
func (c *Client) GetIssueFromGithubID(ghIssueID int64) (*jiralib.Issue, error) {
	issues, _, err := c.JiraClient.Issue.Search(fmt.Sprintf("project = '%s' AND 'GitHub ID' = %d", c.ProjectKey, ghIssueID), &jiralib.SearchOptions{
		MaxResults: 1,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get Jira issue with GitHub ID %d", ghIssueID)
	}
	if len(issues) == 0 {
		return nil, nil
	}
	return &issues[0], nil
}

// UpdateIssue will update a given issue.
//
// JIRA API docs: https://developer.atlassian.com/cloud/jira/platform/rest/v3/?utm_source=/cloud/jira/platform/rest/&utm_medium=302#api-api-3-issue-id-put
func (c *Client) UpdateIssue(issue *jiralib.Issue) (*jiralib.Issue, error) {
	if issue.Fields == nil {
		issue.Fields = &jiralib.IssueFields{
			Unknowns: make(map[string]interface{}),
		}
	} else if issue.Fields.Unknowns == nil {
		issue.Fields.Unknowns = make(map[string]interface{})
	}
	issue.Fields.Unknowns[c.GetCustomFieldID(CFNameGitHubLastIssueSync)] = time.Now().Format(issueSyncDateFormat)
	issue, _, err := c.JiraClient.Issue.Update(issue)
	return issue, errors.Wrap(err, "failed to update issue")
}

// CreateIssue creates an issue or a sub-task from a JSON representation.
//
// JIRA API docs: https://docs.atlassian.com/jira/REST/latest/#api/2/issue-createIssues
func (c *Client) CreateIssue(issueType, status, summary, description, epicKey string, sprint *int, githubID int64, githubNumber int, githubLabels []string, githubStatus string) (*jiralib.Issue, error) {

	issue := &jiralib.Issue{
		Fields: &jiralib.IssueFields{
			Type: jiralib.IssueType{
				Name: issueType,
			},
			Project: jiralib.Project{
				Key: c.ProjectKey,
			},
			Summary:     summary,
			Description: description,
			Unknowns: map[string]interface{}{
				c.customFieldsIDs[CFNameGitHubID]:            githubID,
				c.customFieldsIDs[CFNameGitHubNumber]:        githubNumber,
				c.customFieldsIDs[CFNameGitHubStatus]:        githubStatus,
				c.customFieldsIDs[CFNameGitHubLabels]:        strings.Join(githubLabels, " "),
				c.customFieldsIDs[CFNameGitHubLastIssueSync]: time.Now().Format(issueSyncDateFormat),
			},
		},
	}

	if issueType == "Epic" {
		issue.Fields.Unknowns[c.customFieldsIDs[CFNameEpicName]] = summary
	}
	if epicKey != "" {
		issue.Fields.Unknowns[c.customFieldsIDs[CFNameEpicLink]] = epicKey
	}
	if sprint != nil {
		issue.Fields.Unknowns[c.customFieldsIDs[CFNameSprint]] = *sprint
	}

	issue, resp, err := c.JiraClient.Issue.Create(issue)
	if err != nil && resp != nil {
		defer resp.Body.Close()
		b, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Jira Issue creation error: %s", b)
	}
	return issue, errors.Wrap(err, "failed to create issue")
}

// UpdateIssueEstimate estimate for the given issue key or ID
func (c *Client) UpdateIssueEstimate(issueKeyOrID string, estimate float32) error {
	var estimation struct {
		Value float32 `json:"value,omitempty"`
	}
	estimation.Value = estimate
	req, err := c.JiraClient.NewRequest("PUT", fmt.Sprintf("/rest/agile/1.0/issue/%s/estimation?boardId=%d", issueKeyOrID, c.BoardID), &estimation)
	if err != nil {
		return errors.Wrapf(err, "estimate issue %q", issueKeyOrID)
	}

	resp, err := c.JiraClient.Do(req, nil)
	if err != nil {
		err = jiralib.NewJiraError(resp, err)
		return errors.Wrapf(err, "estimate issue %q", issueKeyOrID)
	}
	return nil
}

// GetIssueEstimate get the estimate for the given issue key or ID
func (c *Client) GetIssueEstimate(issueKeyOrID string) (float32, error) {

	req, err := c.JiraClient.NewRequest("GET", fmt.Sprintf("/rest/agile/1.0/issue/%s/estimation?boardId=%d", issueKeyOrID, c.BoardID), nil)
	if err != nil {
		return 0, errors.Wrapf(err, "estimate issue %q", issueKeyOrID)
	}
	var estimation struct {
		Value float32 `json:"value,omitempty"`
	}
	resp, err := c.JiraClient.Do(req, &estimation)
	if err != nil {
		err = jiralib.NewJiraError(resp, err)
		return 0, errors.Wrapf(err, "estimate issue %q", issueKeyOrID)
	}
	return estimation.Value, nil
}
