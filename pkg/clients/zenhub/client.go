package zenhub

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/google/go-github/v24/github"
	"github.com/pkg/errors"
)

// API abstracts JIRA API to things needed by this project
// This is useful for mocking.
type API interface {
	// GetMilestoneStartDate returns the start date for a given milestone if any.
	//
	// If there is no start date defined the nil is returned.
	// ZenHub API docs: https://github.com/ZenHubIO/API#get-milestone-start-date
	GetMilestoneStartDate(milestoneNumber int) (*time.Time, error)
	// DecorateGHMilestone transforms a GitHub Milestone into a ZenHub Milestone.
	DecorateGHMilestone(ghMilestone *github.Milestone) (*Milestone, error)
}

// Client manages communication with the ZenHub API.
type Client struct {
	BaseURL    *url.URL
	AuthToken  string
	UserAgent  string
	Repository int64
	Verbose    bool
}

const (
	defaultBaseURL    = "https://api.zenhub.io/"
	defaultUserAgent  = "zenhub-client"
	apiRequestTimeout = 30 * time.Second
)

// NewClient creates a client with the given token on the given repository,
// will use defaults for everything else.
//
// See NewClientWithOptions to customize defaults
func NewClient(authToken string, repo int64) *Client {
	u, _ := url.Parse(defaultBaseURL)
	return &Client{
		BaseURL:    u,
		AuthToken:  authToken,
		UserAgent:  defaultUserAgent,
		Verbose:    false,
		Repository: repo,
	}
}

// NewClientWithOptions creates a client with the given token on the given repository
// and let custumize the base URL and verbosity.
func NewClientWithOptions(authToken string, repo int64, baseurl string, verbose bool) (*Client, error) {
	u, err := url.Parse(baseurl)
	if err != nil {
		return nil, err
	}
	return &Client{
		BaseURL:    u,
		AuthToken:  authToken,
		UserAgent:  defaultUserAgent,
		Repository: repo,
		Verbose:    verbose,
	}, nil
}

func (c *Client) urlFor(params ...string) *url.URL {
	newURL, err := url.Parse(c.BaseURL.String())
	if err != nil {
		panic("invalid BaseURL passed")
	}

	if len(params) <= 0 {
		panic("too few arguments")
	}

	newURL.Path = params[0]
	if len(params) >= 2 && params[1] != "" {
		newURL.RawQuery = params[1]
	}
	return newURL
}

func (c *Client) buildReq(req *http.Request) *http.Request {
	req.Header.Set("x-authentication-token", c.AuthToken)
	req.Header.Set("User-Agent", c.UserAgent)
	if req.Method == "PUT" || req.Method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return req
}

// Request sends an API request and returns the API response.
func (c *Client) Request(req *http.Request) (resp *http.Response, err error) {
	req = c.buildReq(req)

	if c.Verbose {
		dump, err := httputil.DumpRequest(req, true)
		if err == nil {
			log.Printf("%s", dump)
		}
	}

	client := &http.Client{}
	client.Timeout = apiRequestTimeout
	resp, err = client.Do(req)

	if err != nil {
		return nil, err
	}
	if c.Verbose {
		dump, err := httputil.DumpResponse(resp, true)
		if err == nil {
			log.Printf("%s", dump)
		}
	}

	if resp.StatusCode/100 != 2 {
		return nil, errors.New(fmt.Sprintf("request failed. status code is %d", resp.StatusCode))
	}

	return resp, nil
}
