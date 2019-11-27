package zenhub

import (
	"time"

	gh "github.com/google/go-github/v24/github"
)

// Milestone represents a ZenHub Milestone.
// It supersede a GitHub Milestone by adding a Start date
type Milestone struct {
	*gh.Milestone `json:"-"`
	StartDate     *time.Time `json:"start_date,omitempty"`
}

// ReleaseReport represents a ZenHub Release report.
type ReleaseReport struct {
	ID             string     `json:"release_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	DesiredEndDate *time.Time `json:"desired_end_date,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
	ClosedAt       *time.Time `json:"closed_at,omitempty"`
	State          string     `json:"state"`
}

// IssueID represents a ZenHub Issue identifier.
type IssueID struct {
	IssueNumber *int `json:"issue_number,omitempty"`
	// When coming from an epic or a release report
	RepoID *int64 `json:"repo_id,omitempty"`
}

// Issue represents a ZenHub Issue.
// It supersede a GitHub Issue by adding a estimate, epic flag, position and pipeline
type Issue struct {
	IssueID
	*gh.Issue `json:"-"`
	IsEpic    bool               `json:"is_epic,omitempty"`
	Estimate  *Estimate          `json:"estimate,omitempty"`
	Position  *int               `json:"position,omitempty"`
	Pipeline  *IssueDataPipeline `json:"pipeline,omitempty"`
}

// Estimate represents an Issue estimate
type Estimate struct {
	Value int `json:"value"`
}

// IssueDataPipeline represent a pipeline section in the Issue description when getting it directly from the Get Issue Data API.
// This does not make sens when getting Issues from the board API.
type IssueDataPipeline struct {
	Name string `json:"name,omitempty"`
}

// Board represents the root data structure of the Board API
type Board struct {
	Pipelines []BoardPipeline `json:"pipelines,omitempty"`
}

// BoardPipeline represents a pipeline within the board
type BoardPipeline struct {
	ID     string   `json:"id,omitempty"`
	Name   string   `json:"name,omitempty"`
	Issues []*Issue `json:"issues,omitempty"`
}

// Epic represents a ZenHub Epic
type Epic struct {
	*Issue
	TotalEpicEstimates *Estimate `json:"total_epic_estimates,omitempty"`
	Issues             []Issue   `json:"issues,omitempty"`
}
