package jira

import (
	jiralib "github.com/andygrunwald/go-jira"
)

// Client manages communication with the Jira API.
type Client struct {
	JiraClient      *jiralib.Client
	ProjectKey      string
	BoardID         int
	customFieldsIDs map[string]string
}

// Init initialize the client by getting some instance specific IDs
//
// Call to this function is required before using it to synchronize ZenHub and Jira
func (c *Client) Init() error {
	return c.initCustomFields()
}
