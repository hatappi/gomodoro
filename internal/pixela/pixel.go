package pixela

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hatappi/go-kit/log"
)

type postPixelResponse struct {
	Message   string `json:"message"`
	IsSuccess bool   `json:"isSuccess"`
}

// IncrementPixel increments a pixel
func (c *Client) IncrementPixel(ctx context.Context, userName, graphID string) error {
	url := "https://pixe.la/v1/users/" + userName + "/graphs/" + graphID + "/increment"
	req, err := http.NewRequest(
		"PUT",
		url,
		nil,
	)
	if err != nil {
		return err
	}
	log.FromContext(ctx).V(1).Info("request: increment a pixel", "url", url)

	req.Header.Set("X-USER-TOKEN", c.token)

	res, err := c.httpclient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	log.FromContext(ctx).V(1).Info("response: increment a pixel", "response body", string(body))

	var postPixel postPixelResponse
	if err = json.Unmarshal(body, &postPixel); err != nil {
		return err
	}

	if !postPixel.IsSuccess {
		return fmt.Errorf("request failed. detail: %s", postPixel.Message)
	}

	return nil
}
