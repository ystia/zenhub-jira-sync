package pkg

import (
	"context"
	"log"
	"time"

	jiralib "github.com/andygrunwald/go-jira"
	"github.com/ystia/zenhub-jira-sync/pkg/clients/zenhub"
)

func (s *Sync) diffMilestoneAndSprint(m *zenhub.Milestone, sprint *jiralib.Sprint) bool {
	updateSprint := false
	switch *m.State {
	case "open":
		if sprint.State == "close" || sprint.State == "future" && (*m.StartDate).Before(time.Now()) {
			updateSprint = true
			sprint.State = "active"
		}
	case "closed":
		if sprint.State != "closed" {
			updateSprint = true
			sprint.State = "closed"
		}
	}
	if m.StartDate != nil {
		if sprint.StartDate == nil || !(*m.StartDate).Equal(*sprint.StartDate) {
			updateSprint = true
			*sprint.StartDate = *m.StartDate
		}
	}

	if m.DueOn != nil {
		if sprint.EndDate == nil || !(*m.DueOn).Equal(*sprint.EndDate) {
			updateSprint = true
			*sprint.EndDate = *m.DueOn
		}
	}

	if m.ClosedAt != nil {
		if sprint.CompleteDate == nil || !(*m.ClosedAt).Equal(*sprint.CompleteDate) {
			updateSprint = true
			*sprint.CompleteDate = *m.ClosedAt
		}
	}
	return updateSprint
}

func (s *Sync) milestones(ctx context.Context) error {
	log.Print("Listing milestones")
	ghMilestones, err := s.GithubClient.ListMilestones(ctx)
	if err != nil {
		return err
	}

	for _, m := range ghMilestones {
		log.Printf("Found GitHub milestone %q, state %q", *m.Title, *m.State)
	}

	log.Print("Listing Spints")
	jiraSprints, err := s.JiraClient.ListSprints(ctx)
	if err != nil {
		return err
	}

	for _, s := range jiraSprints {
		log.Printf("Found Jira sprint %q, state %q", s.Name, s.State)
	}

	log.Print("Comparing milestones and sprints")
	for _, m := range ghMilestones {
		msExists := false
		milestone, err := s.ZenhubClient.DecorateGHMilestone(m)
		if err != nil {
			return err
		}
		for _, sprint := range jiraSprints {
			if *m.Title == sprint.Name {
				msExists = true
				if s.diffMilestoneAndSprint(milestone, &sprint) {
					s.JiraClient.UpdateSprint(&sprint)
				}
			}
		}
		// Create only new milestones
		if !msExists && *milestone.State != "closed" {
			sprint, err := s.JiraClient.CreateSprint(*milestone.Title, "", milestone.StartDate, milestone.DueOn)
			if err != nil {
				return err
			}
			if milestone.StartDate != nil && (*milestone.StartDate).Before(time.Now()) {
				sprint.State = "active"
				s.JiraClient.UpdateSprint(sprint)
			}
		}

	}

	return nil
}
