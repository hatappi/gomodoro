package toggl

import (
	"net/http"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client manage toggl.
type Client struct {
	projectID   int
	workspaceID int
	apiToken    string

	httpclient httpClient
}

// NewClient initilize toggl client.
func NewClient(projectID int, workspaceID int, apiToken string) *Client {
	return &Client{
		projectID:   projectID,
		workspaceID: workspaceID,
		apiToken:    apiToken,
		httpclient:  &http.Client{},
	}
}
