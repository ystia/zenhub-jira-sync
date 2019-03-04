package github

import (
	"testing"
)

func TestAPIImplementation(t *testing.T) {
	var _ API = (*Client)(nil)
}
