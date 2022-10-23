package unix

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"

	"github.com/hatappi/go-kit/log"
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
	b, err := io.ReadAll(bufio.NewReader(c.conn))
	if err != nil {
		log.FromContext(ctx).Error(err, "failed to read connection")
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
