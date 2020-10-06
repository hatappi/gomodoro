// Package pixela manage Pixela
package pixela

import (
	"net/http"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client manage toggl
type Client struct {
	token string

	httpclient httpClient
}

// NewClient initilize toggl client
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpclient: &http.Client{},
	}
}
