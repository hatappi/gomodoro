package toggl

import (
	"net/http"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	projectID int
	apiToken  string

	httpclient HttpClient
}

func NewClient(projectID int, apiToken string) *Client {
	return &Client{
		projectID:  projectID,
		apiToken:   apiToken,
		httpclient: &http.Client{},
	}
}
