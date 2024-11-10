// Package toggl manage toggl
package toggl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const appName = "gomodoro"

// TimeEntry time_entry of toggl.
type TimeEntry struct {
	Description string   `json:"description"`
	CreatedWith string   `json:"created_with"`
	Start       string   `json:"start"`
	Duration    int      `json:"duration"`
	ProjectID   int      `json:"project_id"`
	WorkspaceID int      `json:"workspace_id"`
	TAGS        []string `json:"tags"`
}

// PostTimeEntry record duration with description.
func (c *Client) PostTimeEntry(ctx context.Context, desc string, start time.Time, duration int) error {
	timeEntry := &TimeEntry{
		Description: desc,
		CreatedWith: appName,
		Start:       start.Format("2006-01-02T15:04:05Z07:00"),
		Duration:    duration,
		WorkspaceID: c.workspaceID,
		ProjectID:   c.projectID,
	}

	jsonBytes, err := json.Marshal(timeEntry)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("https://api.track.toggl.com/api/v9/workspaces/%d/time_entries", c.workspaceID),
		bytes.NewBuffer(jsonBytes),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.apiToken, "api_token")

	res, err := c.httpclient.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read body: %w", err)
		}
		return fmt.Errorf("request failed. status: %d, detail: %s", res.StatusCode, body)
	}
	return nil
}
