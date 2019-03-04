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
