package jira

import (
	"strconv"

	"github.com/pkg/errors"
)

func (c *Client) GetProjectID() (int, error) {
	p, _, err := c.JiraClient.Project.Get(c.ProjectKey)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to get project %q", c.ProjectKey)
	}
	id, err := strconv.Atoi(p.ID)

	return id, errors.Wrap(err, "failed to convert returned project ID into a integer")
}
