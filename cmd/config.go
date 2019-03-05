package cmd

import (
	"github.com/pkg/errors"
)

// Config represents a ZenHub To Jira Configuration
type Config struct {
	Synchronizations   []Synchronization  `mapstructure:"synchronizations"`
	JiraURI            string             `mapstructure:"jira_uri"`
	JiraProject        string             `mapstructure:"jira_project"`
	JiraAuthentication JiraAuthentication `mapstructure:"jira_authentication"`
	ZenhubAPIToken     string             `mapstructure:"zenhub_api_token"`
	GithubAPIToken     string             `mapstructure:"github_api_token"`
}

// Synchronization allows to link specific github repository to a Jira Board
type Synchronization struct {
	GithubOwner      string         `mapstructure:"github_owner"`
	GithubRepository string         `mapstructure:"github_repository"`
	JiraBoardID      int            `mapstructure:"jira_board_id"`
	ReleaseRenamer   ReleaseRenamer `mapstructure:"release_renamer"`
}

type ReleaseRenamer struct {
	Source string
	Target string
}

// JiraAuthentication defines how to connect to Jira
type JiraAuthentication struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

func validateConfig(cfg *Config) error {
	if cfg.JiraURI == "" {
		return errors.New("missing jira_uri parameter")
	}
	if cfg.JiraProject == "" {
		return errors.New("missing jira_project parameter")
	}
	if cfg.ZenhubAPIToken == "" {
		return errors.New("missing zenhub_api_token parameter")
	}
	if cfg.GithubAPIToken == "" {
		return errors.New("missing github_api_token parameter")
	}

	if cfg.JiraAuthentication.User == "" {
		return errors.New("missing jira_authentication.user parameter")
	}
	if cfg.JiraAuthentication.Password == "" {
		return errors.New("missing jira_authentication.password parameter")
	}

	for i, s := range cfg.Synchronizations {
		if s.GithubOwner == "" {
			return errors.Errorf("missing jira_authentication[%d].github_owner parameter", i)
		}
		if s.GithubRepository == "" {
			return errors.Errorf("missing jira_authentication[%d].github_repository parameter", i)
		}
		if s.JiraBoardID == 0 {
			return errors.Errorf("missing jira_authentication[%d].jira_board_id parameter", i)
		}

	}

	return nil
}
