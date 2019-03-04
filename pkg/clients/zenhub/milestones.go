package zenhub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/go-github/v24/github"
	"github.com/pkg/errors"
)

// GetMilestoneStartDate returns the start date for a given milestone if any.
//
// If there is no start date defined the nil is returned.
// ZenHUb API docs: https://github.com/ZenHubIO/API#get-milestone-start-date
func (c *Client) GetMilestoneStartDate(milestoneNumber int) (*time.Time, error) {

	req, err := http.NewRequest("GET", c.urlFor(fmt.Sprintf("/p1/repositories/%d/milestones/%d/start_date", c.Repository, milestoneNumber)).String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create zenhub request to get milestone start date")
	}

	resp, err := c.Request(req)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to execute zenhub request to get milestone start date")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read zenhub response to get milestone start date")
	}

	var startDateResp struct {
		StartDate *time.Time `json:"start_date,omitempty"`
	}
	err = json.Unmarshal(body, &startDateResp)

	return startDateResp.StartDate, errors.Wrap(err, "Failed to read zenhub response to get milestone start date")

}

// DecorateGHMilestone transforms a GitHub Milestone into a ZenHub Milestone.
func (c *Client) DecorateGHMilestone(ghMilestone *github.Milestone) (*Milestone, error) {
	startDate, err := c.GetMilestoneStartDate(*ghMilestone.Number)
	return &Milestone{Milestone: ghMilestone, StartDate: startDate}, err
}
