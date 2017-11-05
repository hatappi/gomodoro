package toggl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hatappi/gomodoro/libs/toggl/config"
)

const AppName = "gomodoro"

type TimeEntry struct {
	Description string   `json:"description"`
	CreatedWith string   `json:"created_with"`
	Start       string   `json:"start"`
	Duration    int      `json:"duration"`
	PID         int      `json:"pid"`
	TAGS        []string `json:"tags"`
}

type TimeEntryBody struct {
	TimeEntry *TimeEntry `json:"time_entry"`
}

func PostTimeEntry(conf *toggl.Config, desc string, start time.Time, duration int) error {
	timeEntry := &TimeEntry{
		desc,
		AppName,
		start.Format("2006-01-02T15:04:05Z07:00"),
		duration,
		conf.PID,
		conf.Tags,
	}

	jsonBytes, err := json.Marshal(&TimeEntryBody{
		timeEntry,
	})
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
	req.SetBasicAuth(conf.ApiToken, "api_token")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("request failed. detail: %s", res.Body)
	}
	return nil
}
