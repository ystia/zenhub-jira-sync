package pkg

import (
	"context"
	"fmt"
	"strings"

	jiralib "github.com/andygrunwald/go-jira"
	gh "github.com/google/go-github/v24/github"
)

func (s *Sync) compareComments(ctx context.Context, ghIssue *gh.Issue, jiraIssue *jiralib.Issue) error {

	if ghIssue.GetComments() == 0 {
		// no comments
		return nil
	}

	ghComments, err := s.GithubClient.GetIssueComments(ctx, ghIssue.GetNumber())
	if err != nil {
		return err
	}

	var jiraComments []*jiralib.Comment
	if jiraIssue.Fields != nil && jiraIssue.Fields.Comments != nil {
		jiraComments = jiraIssue.Fields.Comments.Comments
	}

	for _, ghc := range ghComments {
		var commentFound bool
		for _, jc := range jiraComments {
			if strings.Contains(jc.Body, fmt.Sprintf("GitHub Comment: ID: [%d]", ghc.GetID())) {
				commentFound = true
				if !strings.Contains(jc.Body, ghc.GetBody()) {
					err := s.updateJiraComment(ghc, jiraIssue, jc)
					if err != nil {
						return err
					}
				}
				break
			}
		}
		if !commentFound {
			err := s.createJiraComment(ghc, jiraIssue)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getJiraCommentBodyFromGHComment(ghc *gh.IssueComment) string {
	return fmt.Sprintf("GitHub Comment: ID: [%d], User: [%s]\n\n---------------------\n\n%s", ghc.GetID(), ghc.GetUser().GetLogin(), ghc.GetBody())
}

func (s *Sync) createJiraComment(ghComment *gh.IssueComment, jiraIssue *jiralib.Issue) error {
	_, err := s.JiraClient.AddComment(jiraIssue.Key, getJiraCommentBodyFromGHComment(ghComment))
	return err
}

func (s *Sync) updateJiraComment(ghComment *gh.IssueComment, jiraIssue *jiralib.Issue, jiraComment *jiralib.Comment) error {
	_, err := s.JiraClient.UpdateComment(jiraIssue.Key, jiraComment.ID, getJiraCommentBodyFromGHComment(ghComment))
	return err
}
