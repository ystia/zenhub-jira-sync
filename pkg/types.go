package pkg

import (
	"context"
	"regexp"

	"github.com/ystia/zenhub-jira-sync/pkg/clients/github"
	"github.com/ystia/zenhub-jira-sync/pkg/clients/jira"
	"github.com/ystia/zenhub-jira-sync/pkg/clients/zenhub"
)

// Sync is our synchronization tool
type Sync struct {
	GithubClient      github.API
	JiraClient        jira.API
	ZenhubClient      zenhub.API
	ReleaseNameRE     *regexp.Regexp
	VersionNameRename string
	DefaultIssueType  string
	LabelsToIssueType []struct {
		Label     string
		IssueType string
	}
	DefaultJiraComponents []string
}

// All synchronize every thing
func (s *Sync) All(ctx context.Context) error {

	err := s.milestones(ctx)
	if err != nil {
		return err
	}

	relTuples, err := s.releases()
	if err != nil {
		return err
	}
	return s.issues(ctx, relTuples)
}

type releasesTuple struct {
	zhRelease   *zenhub.ReleaseReport
	jiraVersion *jira.Version
}
