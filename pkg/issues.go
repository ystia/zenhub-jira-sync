package pkg

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	jiralib "github.com/andygrunwald/go-jira"
	gh "github.com/google/go-github/v24/github"

	"github.com/ystia/zenhub-jira-sync/pkg/clients/jira"
	"github.com/ystia/zenhub-jira-sync/pkg/clients/zenhub"
)

var sprintIDRE = regexp.MustCompile(`id=(\d+)`)

func (s *Sync) issues(ctx context.Context) error {
	sprintNamesToIDs := make(map[string]int)

	sprintList, err := s.JiraClient.ListSprints(ctx)
	for _, sprint := range sprintList {
		sprintNamesToIDs[sprint.Name] = sprint.ID
	}

	issuesToEpics := make(map[string]string)

	// First create Epics
	epics, err := s.ZenhubClient.GetEpics()

	for _, epic := range epics {
		jiraEpic, err := s.checkIssue(ctx, epic.Issue, "", sprintNamesToIDs)
		if err != nil {
			return err
		}
		for _, issue := range epic.Issues {
			if issue.IsEpic {
				// Jira doesn't support Epics within epics
				continue
			}
			issuesToEpics[fmt.Sprintf("%d/%d", *issue.RepoID, *issue.IssueNumber)] = jiraEpic.Key
		}
	}

	board, err := s.ZenhubClient.GetBoard()
	if err != nil {
		return err
	}

	for _, pipeline := range board.Pipelines {
		log.Printf("Synchronizing issues from pipeline %q", pipeline.Name)
		for _, issue := range pipeline.Issues {
			if issue.IsEpic {
				// epics already synchronized
				continue
			}
			issue.Pipeline = &zenhub.IssueDataPipeline{Name: pipeline.Name}
			epicKey := issuesToEpics[fmt.Sprintf("%d/%d", *issue.RepoID, *issue.IssueNumber)]
			_, err = s.checkIssue(ctx, issue, epicKey, sprintNamesToIDs)
			if err != nil {
				return err
			}
		}
	}
	// Closed issues do not appear in ZH board
	return s.checkClosedIssues(ctx)
}

func (s *Sync) checkClosedIssues(ctx context.Context) error {
	log.Printf("Checking issues closed on GitHub")
	closedIssues, err := s.GithubClient.ListIssues(ctx, &gh.IssueListByRepoOptions{
		State: "closed",
	})
	if err != nil {
		return err
	}

	for _, issue := range closedIssues {
		jiraIssue, err := s.JiraClient.GetIssueFromGithubID(issue.GetID())
		if err != nil {
			return err
		}
		if jiraIssue == nil {
			continue
		}
		if jiraIssue.Fields.Status != nil && !(jiraIssue.Fields.Status.Name == "Closed" || jiraIssue.Fields.Status.Name == "Done") {
			err = s.JiraClient.TransitionIssue(jiraIssue.Key, "Close Issue")
			if err != nil {
				err = s.JiraClient.TransitionIssue(jiraIssue.Key, "Done")
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *Sync) checkIssue(ctx context.Context, issue *zenhub.Issue, epicKey string, sprintNamesToIDs map[string]int) (*jiralib.Issue, error) {
	ghIssue, err := s.GithubClient.GetIssue(ctx, *issue.IssueNumber)
	if err != nil {
		return nil, err
	}
	issue.Issue = ghIssue

	jiraIssue, err := s.JiraClient.GetIssueFromGithubID(*ghIssue.ID)
	if err != nil {
		return nil, err
	}
	if jiraIssue != nil {
		jiraIssueUpdate, changed, moveToBacklog, updateEstimate := s.diffIssues(issue, jiraIssue, epicKey, sprintNamesToIDs)
		if changed {
			jiraIssue, err = s.JiraClient.UpdateIssue(jiraIssueUpdate)
			if err != nil {
				return nil, err
			}
		}
		if moveToBacklog {
			err = s.JiraClient.MoveToBacklog([]string{jiraIssue.Key})
			if err != nil {
				return nil, err
			}
		}

		if updateEstimate {
			var estimate int
			if issue.Estimate != nil {
				estimate = issue.Estimate.Value
			}
			err = s.JiraClient.UpdateIssueEstimate(jiraIssue.Key, float32(estimate))
			if err != nil {
				return nil, err
			}
		}

		err = s.checkContainsRemoteURL(jiraIssue, "Original GitHub Issue", issue.GetHTMLURL())
		if err != nil {
			return nil, err
		}
	} else {
		jiraIssue, err = s.createJiraIssueFromZenHubIssue(issue, epicKey, sprintNamesToIDs)
		if err != nil {
			return nil, err
		}
	}
	err = s.compareComments(ctx, ghIssue, jiraIssue)
	return jiraIssue, err
}

func (s *Sync) diffIssues(zhIssue *zenhub.Issue, jiraIssue *jiralib.Issue, epicKey string, sprintNamesToIDs map[string]int) (*jiralib.Issue, bool, bool, bool) {
	var updatedIssue bool
	var moveToBacklog bool
	resultIssue := &jiralib.Issue{
		Key: jiraIssue.Key,
		ID:  jiraIssue.ID,
		Fields: &jiralib.IssueFields{
			Unknowns: map[string]interface{}{},
		},
	}

	if s.checkDefaultComponentsOnIssue(jiraIssue) {
		updatedIssue = true
		resultIssue.Fields.Components = jiraIssue.Fields.Components
	}

	if zhIssue.GetTitle() != jiraIssue.Fields.Summary {
		updatedIssue = true
		resultIssue.Fields.Summary = zhIssue.GetTitle()
	}
	if zhIssue.GetBody() != jiraIssue.Fields.Description {
		updatedIssue = true
		resultIssue.Fields.Description = zhIssue.GetBody()
	}
	zhLabels := strings.Join(getZHIssueLabels(zhIssue), " ")
	var jLabels string
	if v, ok := jiraIssue.Fields.Unknowns[s.JiraClient.GetCustomFieldID(jira.CFNameGitHubLabels)]; ok {
		jLabels, ok = v.(string)
	}
	if jLabels != zhLabels {
		updatedIssue = true
		resultIssue.Fields.Unknowns[s.JiraClient.GetCustomFieldID(jira.CFNameGitHubLabels)] = zhLabels
	}

	sprintRef := jiraIssue.Fields.Unknowns[s.JiraClient.GetCustomFieldID(jira.CFNameSprint)]
	var sprintID int
	if sprintRef != nil {
		matches := sprintIDRE.FindStringSubmatch(fmt.Sprintf("%v", sprintRef))
		if len(matches) >= 2 {
			sprintID, _ = strconv.Atoi(matches[1])
		}
	}
	if zhIssue.Milestone == nil && sprintRef != nil {
		moveToBacklog = true
	} else if zhIssue.Milestone != nil && sprintID != sprintNamesToIDs[zhIssue.Milestone.GetTitle()] {
		updatedIssue = true
		resultIssue.Fields.Unknowns[s.JiraClient.GetCustomFieldID(jira.CFNameSprint)] = sprintNamesToIDs[zhIssue.Milestone.GetTitle()]
	}

	if jiraIssue.Fields.Unknowns[s.JiraClient.GetCustomFieldID(jira.CFNameGitHubStatus)] != zhIssue.GetState() {
		updatedIssue = true
		resultIssue.Fields.Unknowns[s.JiraClient.GetCustomFieldID(jira.CFNameGitHubStatus)] = zhIssue.GetState()
	}

	var updateEstimate bool
	if jira.IsIssueTypeEstimable(jiraIssue.Fields.Type.Name) {
		var estimate int
		jiraSP, err := s.JiraClient.GetIssueEstimate(jiraIssue.ID)
		if err == nil {
			if zhIssue.Estimate != nil {
				estimate = zhIssue.Estimate.Value
			}
			if jiraSP != float32(estimate) {
				updateEstimate = true
			}
		} else {
			log.Printf("failed to get issue estimate %v, do not update it", err)
		}
	}
	return resultIssue, updatedIssue, moveToBacklog, updateEstimate
}

func (s *Sync) checkDefaultComponentsOnIssue(jiraIssue *jiralib.Issue) bool {
	var updated bool
	for _, defComp := range s.DefaultJiraComponents {
		var found bool
		for _, comp := range jiraIssue.Fields.Components {
			if comp.Name == defComp {
				found = true
				break
			}
		}
		if !found {
			updated = true
			jiraIssue.Fields.Components = append(jiraIssue.Fields.Components, &jiralib.Component{Name: defComp})
		}
	}
	return updated
}

func (s *Sync) createJiraIssueFromZenHubIssue(issue *zenhub.Issue, epicKey string, sprintNamesToIDs map[string]int) (*jiralib.Issue, error) {

	issueType := s.DefaultIssueType
	for _, pair := range s.LabelsToIssueType {
		if hasLabel(issue, pair.Label) {
			issueType = pair.IssueType
			break
		}
	}
	var sprint *int
	if issue.GetMilestone() != nil {
		sprint = new(int)
		*sprint = sprintNamesToIDs[issue.GetMilestone().GetTitle()]
	}

	jiraIssue, err := s.JiraClient.CreateIssue(issueType, issue.GetTitle(), issue.GetBody(), epicKey, s.DefaultJiraComponents, sprint, issue.GetID(), issue.GetNumber(), getZHIssueLabels(issue), issue.GetState())
	if err != nil {
		return jiraIssue, err
	}
	if issue.Estimate != nil && jira.IsIssueTypeEstimable(issueType) {
		err = s.JiraClient.UpdateIssueEstimate(jiraIssue.ID, float32(issue.Estimate.Value))
		if err != nil {
			return jiraIssue, err
		}
	}

	err = s.JiraClient.AddRemoteLinkToIssue(jiraIssue.ID, "", "Original GitHub Issue", issue.GetHTMLURL())
	return jiraIssue, err
}

func getZHIssueLabels(issue *zenhub.Issue) []string {
	labels := make([]string, 0)
	for _, l := range issue.Labels {
		labels = append(labels, l.GetName())
	}
	return labels
}

func hasLabel(issue *zenhub.Issue, labelName string) bool {
	for _, l := range issue.Labels {
		if l.Name != nil && *l.Name == labelName {
			return true
		}
	}
	return false
}

func (s *Sync) checkContainsRemoteURL(jiraIssue *jiralib.Issue, title, url string) error {
	links, err := s.JiraClient.GetIssueRemoteLinks(jiraIssue.Key)
	if err != nil {
		return err
	}
	for _, link := range links {
		if link.Object.Title == title && link.Object.URL == url {
			return nil
		}
	}
	return s.JiraClient.AddRemoteLinkToIssue(jiraIssue.Key, "", title, url)
}
