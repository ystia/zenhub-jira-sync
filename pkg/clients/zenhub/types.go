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
