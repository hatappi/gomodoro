package toggl

import (
	"net/http"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client manage toggl.
type Client struct {
	projectID int
	apiToken  string

	httpclient httpClient
}

// NewClient initilize toggl client.
func NewClient(projectID int, apiToken string) *Client {
	return &Client{
		projectID:  projectID,
		apiToken:   apiToken,
		httpclient: &http.Client{},
	}
}
