package pixela

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

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
	log.FromContext(ctx).Debug("request: increment a pixel", zap.String("url", url))

	req.Header.Set("X-USER-TOKEN", c.token)

	res, err := c.httpclient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	log.FromContext(ctx).Debug("response: increment a pixel", zap.String("response body", string(body)))

	var postPixel postPixelResponse
	if err = json.Unmarshal(body, &postPixel); err != nil {
		return err
	}

	if !postPixel.IsSuccess {
		return xerrors.Errorf("request failed. detail: %s", postPixel.Message)
	}

	return nil
}
