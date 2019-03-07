package zenhub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	gh "github.com/google/go-github/v24/github"
	"github.com/pkg/errors"
)

// GetIssue
func (c *Client) GetIssue(issueNumber int) (*Issue, error) {
	req, err := http.NewRequest("GET", c.urlFor(fmt.Sprintf("/p1/repositories/%d/issues/%d", c.Repository, issueNumber)).String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create zenhub request to get issue data")
	}

	resp, err := c.Request(req)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to execute zenhub request to get issue data")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read zenhub response to get issue data")
	}

	issue := new(Issue)
	err = json.Unmarshal(body, issue)

	return issue, errors.Wrap(err, "Failed to read zenhub response to get issue data")
}

func (c *Client) DecorateGithubIssue(ghIssue *gh.Issue) (*Issue, error) {
	issue, err := c.GetIssue(*ghIssue.Number)
	if err != nil {
		return nil, err
	}
	issue.Issue = ghIssue
	return issue, nil
}
