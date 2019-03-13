package jira

import "github.com/pkg/errors"

const (
	CFNameGitHubID            = "GitHub ID"
	CFNameGitHubNumber        = "GitHub Number"
	CFNameGitHubLabels        = "GitHub Labels"
	CFNameGitHubStatus        = "GitHub Status"
	CFNameGitHubReporter      = "GitHub Reporter"
	CFNameGitHubLastIssueSync = "Last Issue-Sync Update"
	CFNameEpicName            = "Epic Name"
	CFNameEpicLink            = "Epic Link"
	CFNameSprint              = "Sprint"
	CFNameStatus              = "Status"
)

func (c *Client) initCustomFields() error {
	c.customFieldsIDs = map[string]string{
		CFNameGitHubID:            "",
		CFNameGitHubNumber:        "",
		CFNameGitHubLabels:        "",
		CFNameGitHubStatus:        "",
		CFNameGitHubReporter:      "",
		CFNameGitHubLastIssueSync: "",
		CFNameEpicName:            "",
		CFNameEpicLink:            "",
		CFNameSprint:              "",
		CFNameStatus:              "",
	}
	jiraFields, _, err := c.JiraClient.Field.GetList()
	if err != nil {
		return errors.Wrap(err, "Failed to get Jira custom fields")
	}

	for _, field := range jiraFields {
		// Get all fields
		c.customFieldsIDs[field.Name] = field.ID
	}

	// Check that pre-registered fields were present
	for cfName, cfID := range c.customFieldsIDs {
		if cfID == "" {
			return errors.Errorf("failed to get ID of field %q in JIRA make sure it have been properly created", cfName)
		}
	}

	return nil
}

// GetCustomFieldID returns a custom field ID based on its name. If not found an empty string is returned.
func (c *Client) GetCustomFieldID(name string) string {
	return c.customFieldsIDs[name]
}

// IsIssueTypeEstimable checks if issue type support estimation through Story Points custom field
//
// Currently only "User story" support estimate
func IsIssueTypeEstimable(issueTypeName string) bool {
	switch issueTypeName {
	case "User story":
		return true
	default:
		return false
	}
}
