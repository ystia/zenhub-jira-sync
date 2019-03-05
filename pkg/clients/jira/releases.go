package jira

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	jiralib "github.com/andygrunwald/go-jira"
	"github.com/pkg/errors"
)

// GetProjectVersions returns all versions of the associated Jira project
func (c *Client) GetProjectVersions() ([]*Version, error) {
	var p struct {
		Versions []*Version `json:"versions,omitempty" structs:"versions,omitempty"`
	}

	req, err := c.JiraClient.NewRequest("GET", fmt.Sprintf("/rest/api/2/project/%s", c.ProjectID), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get versions for Jira project %q", c.ProjectID)
	}

	_, err = c.JiraClient.Do(req, &p)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get versions for Jira project %q", c.ProjectID)
	}

	return p.Versions, nil
}

// CreateVersion creates a version in JIRA.
//
// JIRA API docs: https://developer.atlassian.com/cloud/jira/platform/rest/v2/?utm_source=%2Fcloud%2Fjira%2Fplatform%2Frest%2F&utm_medium=302#api-api-2-version-post
func (c *Client) CreateVersion(name, description string, projectID int, released, archived bool, startDate, dueDate, releaseDate *time.Time) (*Version, error) {

	version := &Version{
		Version: jiralib.Version{
			Name:        name,
			Description: description,
			Released:    released,
			Archived:    archived,
			ProjectID:   projectID,
		},
	}
	if startDate != nil {
		version.StartDate = startDate.Format("2006-01-02")
	}
	if dueDate != nil {
		version.Version.UserReleaseDate = dueDate.Format("2/Jan/2006")
	}
	if releaseDate != nil {
		version.Version.ReleaseDate = releaseDate.Format("2006-01-02")
	}
	if version.ReleaseDate != "" && version.UserReleaseDate != "" {
		// Only one of them could be updated in a single request and user release date
		// is automatically updated to release date when updating release date
		version.UserReleaseDate = ""
	}

	req, err := c.JiraClient.NewRequest("POST", "/rest/api/2/version", version)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create version %q", name)
	}

	resp, err := c.JiraClient.Do(req, version)
	if err != nil {
		defer resp.Body.Close()
		b, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Jira Version creation error: %s", b)
	}
	return version, errors.Wrap(err, "failed to create version")
}

// UpdateVersion will update a given version.
//
// JIRA API docs: https://developer.atlassian.com/cloud/jira/platform/rest/v2/?utm_source=/cloud/jira/platform/rest/&utm_medium=302#api-api-2-version-id-put
func (c *Client) UpdateVersion(version *Version) (*Version, error) {
	if version.ReleaseDate != "" && version.UserReleaseDate != "" {
		// Only one of them could be updated in a single request and user release date
		// is automatically updated to release date when updating release date
		version.UserReleaseDate = ""
	}
	req, err := c.JiraClient.NewRequest("PUT", fmt.Sprintf("/rest/api/2/version/%s", version.ID), version)
	if err != nil {
		return version, errors.Wrapf(err, "failed to update sprint %q", version.Name)
	}
	resp, err := c.JiraClient.Do(req, version)
	if err != nil {
		defer resp.Body.Close()
		b, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Jira Version update error: %s", b)
	}
	return version, errors.Wrapf(err, "failed to update sprint %q", version.Name)
}
