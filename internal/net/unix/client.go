package unix

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"

	"github.com/hatappi/go-kit/log"
)

// Client represents unix server client.
type Client interface {
	Get(context.Context) (*Response, error)

	Close()
}

// IClient implements Client interface.
type IClient struct {
	conn net.Conn
}

// NewClient initialize Client.
func NewClient(socketPath string) (*IClient, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}

	return &IClient{
		conn: conn,
	}, nil
}

// Get a response of connection.
func (c *IClient) Get(ctx context.Context) (*Response, error) {
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

// Close closes client connection.
func (c *IClient) Close() {
	_ = c.conn.Close()
}
