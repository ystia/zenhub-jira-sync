package pkg

import (
	"context"

	"github.com/ystia/zenhub-jira-sync/pkg/clients/zenhub"

	"github.com/ystia/zenhub-jira-sync/pkg/clients/github"
	"github.com/ystia/zenhub-jira-sync/pkg/clients/jira"
)

// Sync is our synchronization tool
type Sync struct {
	GithubClient github.API
	JiraClient   jira.API
	ZenhubClient zenhub.API
}

// All synchronize every thing
func (s *Sync) All(ctx context.Context) error {

	return s.milestones(ctx)

}
