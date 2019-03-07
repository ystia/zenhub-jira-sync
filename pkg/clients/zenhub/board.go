package zenhub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// GetBoard retrieves ZenHub Board
func (c *Client) GetBoard() (*Board, error) {

	req, err := http.NewRequest("GET", c.urlFor(fmt.Sprintf("/p1/repositories/%d/board", c.Repository)).String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create zenhub request to get board")
	}

	resp, err := c.Request(req)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to execute zenhub request to get board")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read zenhub response to get board")
	}

	board := new(Board)
	err = json.Unmarshal(body, board)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read zenhub response to get board")
	}
	for _, pipeline := range board.Pipelines {
		for _, issue := range pipeline.Issues {
			issue.RepoID = &c.Repository
		}
	}
	return board, nil
}
