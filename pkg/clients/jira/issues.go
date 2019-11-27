package jira

import (
	"fmt"
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
		Fields:     []string{"*all"},
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
	issueKey := issue.Key
	issue.Fields.Unknowns[c.GetCustomFieldID(CFNameGitHubLastIssueSync)] = time.Now().Format(issueSyncDateFormat)
	issue, _, err := c.JiraClient.Issue.Update(issue)
	return issue, errors.Wrapf(err, "failed to update issue %q", issueKey)
}

// CreateIssue creates an issue or a sub-task from a JSON representation.
//
// JIRA API docs: https://docs.atlassian.com/jira/REST/latest/#api/2/issue-createIssues
func (c *Client) CreateIssue(issueType, summary, description, epicKey string, components []string, sprint *int, githubID int64, githubNumber int, githubLabels []string, githubStatus string) (*jiralib.Issue, error) {

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

	jiraComponents := make([]*jiralib.Component, len(components))
	for i, compName := range components {
		jiraComponents[i] = &jiralib.Component{
			Name: compName,
		}
	}
	issue.Fields.Components = jiraComponents

	issue, resp, err := c.JiraClient.Issue.Create(issue)
	if err != nil {
		err = jiralib.NewJiraError(resp, err)
		return nil, errors.Wrapf(err, "failed to create issue GH-%d", githubNumber)
	}
	return issue, nil
}

// UpdateIssueEstimate estimate for the given issue key or ID
func (c *Client) UpdateIssueEstimate(issueKeyOrID string, estimate float32) error {
	var estimation struct {
		Value float32 `json:"value,omitempty"`
	}
	estimation.Value = estimate
	req, err := c.JiraClient.NewRequest("PUT", fmt.Sprintf("/rest/agile/1.0/issue/%s/estimation?boardId=%d", issueKeyOrID, c.BoardID), &estimation)
	if err != nil {
		return errors.Wrapf(err, "failed to estimate issue %q", issueKeyOrID)
	}

	resp, err := c.JiraClient.Do(req, nil)
	if err != nil {
		err = jiralib.NewJiraError(resp, err)
		return errors.Wrapf(err, "failed to estimate issue %q", issueKeyOrID)
	}
	return nil
}

// GetIssueEstimate get the estimate for the given issue key or ID
func (c *Client) GetIssueEstimate(issueKeyOrID string) (float32, error) {

	req, err := c.JiraClient.NewRequest("GET", fmt.Sprintf("/rest/agile/1.0/issue/%s/estimation?boardId=%d", issueKeyOrID, c.BoardID), nil)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to get estimate for issue %q", issueKeyOrID)
	}
	var estimation struct {
		Value float32 `json:"value,omitempty"`
	}
	resp, err := c.JiraClient.Do(req, &estimation)
	if err != nil {
		err = jiralib.NewJiraError(resp, err)
		return 0, errors.Wrapf(err, "failed to get estimate for issue %q", issueKeyOrID)
	}
	return estimation.Value, nil
}

// AddRemoteLinkToIssue adds a link to an issue
//
// JIRA API docs: https://developer.atlassian.com/cloud/jira/platform/rest/v3/?utm_source=/cloud/jira/platform/rest/&utm_medium=302#api-rest-api-3-issue-issueIdOrKey-remotelink-post
func (c *Client) AddRemoteLinkToIssue(issueKeyOrID, globalID, title, url string) error {

	remoteLink := RemoteLink{
		GlobalID: globalID,
		Object: RemoteLinkObject{
			Title: title,
			URL:   url,
		},
	}
	req, err := c.JiraClient.NewRequest("POST", fmt.Sprintf("/rest/api/2/issue/%s/remotelink", issueKeyOrID), &remoteLink)
	if err != nil {
		return errors.Wrapf(err, "failed to add remote link for issue issue %q", issueKeyOrID)
	}

	resp, err := c.JiraClient.Do(req, nil)
	if err != nil {
		err = jiralib.NewJiraError(resp, err)
		return errors.Wrapf(err, "failed to add remote link for issue issue %q", issueKeyOrID)
	}
	return nil
}

// GetIssueRemoteLinks get the list of remote links for a given issue
//
// JIRA API docs: https://developer.atlassian.com/cloud/jira/platform/rest/v3/?utm_source=/cloud/jira/platform/rest/&utm_medium=302#api-rest-api-3-issue-issueIdOrKey-remotelink-get
func (c *Client) GetIssueRemoteLinks(issueKeyOrID string) ([]RemoteLink, error) {
	req, err := c.JiraClient.NewRequest("GET", fmt.Sprintf("/rest/api/2/issue/%s/remotelink", issueKeyOrID), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get remote links for issue issue %q", issueKeyOrID)
	}
	remoteLinks := make([]RemoteLink, 0)
	resp, err := c.JiraClient.Do(req, &remoteLinks)
	if err != nil {
		err = jiralib.NewJiraError(resp, err)
		return nil, errors.Wrapf(err, "failed to get remotes link for issue issue %q", issueKeyOrID)
	}
	return remoteLinks, nil
}

// TransitionIssue execute transition identified by the given name to the issue
func (c *Client) TransitionIssue(issueKeyOrID, transitionName string) error {

	transitions, _, err := c.JiraClient.Issue.GetTransitions(issueKeyOrID)
	if err != nil {
		return errors.Wrapf(err, "failed to get transitions for issue %q", issueKeyOrID)
	}

	for _, transition := range transitions {
		if transition.Name == transitionName {
			_, err = c.JiraClient.Issue.DoTransition(issueKeyOrID, transition.ID)
			return errors.Wrapf(err, "failed to apply transition %q to issue %q", transition.Name, issueKeyOrID)
		}
	}

	return errors.Errorf("transition %q not supported on issue %q", transitionName, issueKeyOrID)
}

// UpdateIssueFixVersion will set the fixVersion to the given list of version ids.
//
// JIRA API docs: https://developer.atlassian.com/cloud/jira/platform/rest/v3/?utm_source=/cloud/jira/platform/rest/&utm_medium=302#api-api-3-version-id-put
func (c *Client) UpdateIssueFixVersion(issueKeyOrID string, versionsIDs []string) error {

	type setID struct {
		ID string `json:"id"`
	}
	type setAction struct {
		Set []setID `json:"set"`
	}
	type updateFixVersions struct {
		FixVersions []setAction `json:"fixVersions"`
	}
	type reqBodyStruct struct {
		Update updateFixVersions `json:"update"`
	}
	reqBody := reqBodyStruct{
		Update: updateFixVersions{
			FixVersions: []setAction{
				setAction{
					Set: make([]setID, len(versionsIDs)),
				},
			},
		},
	}
	for i, vid := range versionsIDs {
		reqBody.Update.FixVersions[0].Set[i] = setID{
			ID: vid,
		}
	}

	req, err := c.JiraClient.NewRequest("PUT", fmt.Sprintf("/rest/api/2/issue/%s", issueKeyOrID), &reqBody)
	if err != nil {
		return errors.Wrapf(err, "failed update FixVersions for issue issue %q", issueKeyOrID)
	}
	resp, err := c.JiraClient.Do(req, nil)
	if err != nil {
		err = jiralib.NewJiraError(resp, err)
		return errors.Wrapf(err, "failed update FixVersions for issue issue %q", issueKeyOrID)
	}
	return nil
}
