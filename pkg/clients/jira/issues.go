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
			Labels:      githubLabels,
			Unknowns: map[string]interface{}{
				c.customFieldsIDs[CFNameStatus]:              status,
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
