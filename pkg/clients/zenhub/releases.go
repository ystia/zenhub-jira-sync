package zenhub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// GetReleasesReports returns releases reports for the associated repository.
func (c *Client) GetReleasesReports() ([]*ReleaseReport, error) {
	req, err := http.NewRequest("GET", c.urlFor(fmt.Sprintf("/p1/repositories/%d/reports/releases", c.Repository)).String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create zenhub request to get releases reports")
	}

	resp, err := c.Request(req)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to execute zenhub request to get releases reports")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read zenhub response to get releases reports")
	}

	var res []*ReleaseReport
	err = json.Unmarshal(body, &res)

	return res, errors.Wrap(err, "Failed to read zenhub response to get releases reports")
}

// GetIssuesForReleaseReport returns issues related to a release report
func (c *Client) GetIssuesForReleaseReport(releaseID string) ([]IssueID, error) {
	req, err := http.NewRequest("GET", c.urlFor(fmt.Sprintf("/p1/reports/release/%s/issues", releaseID)).String(), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get zenhub issues for release report %q", releaseID)
	}

	resp, err := c.Request(req)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to execute zenhub request to get issues for release report %q", releaseID)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read zenhub response to get issues for release report %q", releaseID)
	}

	var res []IssueID
	err = json.Unmarshal(body, &res)

	filteredIssues := res[:0]
	for _, i := range res {
		// filter issues that are not from the same repository
		if i.RepoID == nil || *i.RepoID == c.Repository {
			filteredIssues = append(filteredIssues, i)
		}
	}

	return filteredIssues, errors.Wrapf(err, "Failed to read zenhub response to get issues for release report %q", releaseID)
}
