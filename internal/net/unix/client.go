package unix

import (
	"bufio"
	"context"
	"encoding/json"
	"io/ioutil"
	"net"

	"github.com/hatappi/go-kit/log"
	"go.uber.org/zap"
)

// Client represents unix server client
type Client interface {
	Get(context.Context) (*Response, error)

	Close()
}

type clientImpl struct {
	conn net.Conn
}

// NewClient initialize Client
func NewClient(socketPath string) (Client, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}

	return &clientImpl{
		conn: conn,
	}, nil
}

func (c *clientImpl) Get(ctx context.Context) (*Response, error) {
	b, err := ioutil.ReadAll(bufio.NewReader(c.conn))
	if err != nil {
		log.FromContext(ctx).Error("failed to read connection", zap.Error(err))
	}

	r := &Response{}
	if err = json.Unmarshal(b, r); err != nil {
		return nil, err
	}

	return r, nil
}

func (c *clientImpl) Close() {
	_ = c.conn.Close()
}
