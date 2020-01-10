// Package toggl manage toggl
package toggl

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/xerrors"
)

const appName = "gomodoro"

// TimeEntry time_entry of toggl
type TimeEntry struct {
	Description string   `json:"description"`
	CreatedWith string   `json:"created_with"`
	Start       string   `json:"start"`
	Duration    int      `json:"duration"`
	PID         int      `json:"pid"`
	TAGS        []string `json:"tags"`
}

// PostTimeEntry record duration with description
func (c *Client) PostTimeEntry(desc string, start time.Time, duration int) error {
	timeEntry := &TimeEntry{
		Description: desc,
		CreatedWith: appName,
		Start:       start.Format("2006-01-02T15:04:05Z07:00"),
		Duration:    duration,
		PID:         c.projectID,
	}

	body := &struct {
		TimeEntry *TimeEntry `json:"time_entry"`
	}{
		TimeEntry: timeEntry,
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(
		"POST",
		"https://www.toggl.com/api/v8/time_entries",
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

	if res.StatusCode != 200 {
		return xerrors.Errorf("request failed. detail: %s", res.Body)
	}
	return nil
}
