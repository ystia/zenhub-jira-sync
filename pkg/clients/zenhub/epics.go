package zenhub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// GetEpics returns Epics defined in this repository
//
// Github associated issue are not initialized, neither in epics nor in issues
func (c *Client) GetEpics() ([]*Epic, error) {
	req, err := http.NewRequest("GET", c.urlFor(fmt.Sprintf("/p1/repositories/%d/epics", c.Repository)).String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create zenhub request to get epics")
	}

	resp, err := c.Request(req)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to execute zenhub request to get epics")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read zenhub response to get epics")
	}

	type epicList struct {
		EpicIssues []struct {
			IssueNumber int    `json:"issue_number,omitempty"`
			RepoID      *int64 `json:"repo_id,omitempty"`
			IssueURL    string `json:"issue_url,omitempty"`
		} `json:"epic_issues,omitempty"`
	}
	el := new(epicList)
	err = json.Unmarshal(body, el)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode zenhub response to get epics")
	}
	result := make([]*Epic, 0, len(el.EpicIssues))
	for _, ep := range el.EpicIssues {
		epic, err := c.GetEpic(ep.IssueNumber)
		epic.RepoID = ep.RepoID
		if err != nil {
			return nil, err
		}
		result = append(result, epic)
	}
	return result, nil
}

// GetEpic returns a single Epic
//
// Associated issues are filtered to only those that are from the same repository
// Github associated issue are not initialized, neither in epics nor in issues
func (c *Client) GetEpic(epicNumber int) (*Epic, error) {
	req, err := http.NewRequest("GET", c.urlFor(fmt.Sprintf("/p1/repositories/%d/epics/%d", c.Repository, epicNumber)).String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create zenhub request to get epic data")
	}

	resp, err := c.Request(req)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to execute zenhub request to get epic data")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read zenhub response to get epic data")
	}

	epic := new(Epic)
	err = json.Unmarshal(body, epic)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read zenhub response to get epic data")
	}

	filteredIssues := epic.Issues[:0]
	for _, i := range epic.Issues {
		// filter issues that are not from the same repository
		// Also filter of epics
		if i.RepoID == nil || *i.RepoID == c.Repository {
			filteredIssues = append(filteredIssues, i)
		}
	}
	epic.Issues = filteredIssues
	epic.IssueNumber = &epicNumber
	epic.IsEpic = true
	return epic, nil
}
