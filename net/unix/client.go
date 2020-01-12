package unix

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"net"

	"github.com/hatappi/gomodoro/logger"
)

// Client represents unix server client
type Client interface {
	Get() (*Response, error)

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

func (c *clientImpl) Get() (*Response, error) {
	b, err := ioutil.ReadAll(bufio.NewReader(c.conn))
	if err != nil {
		logger.Errorf("error is %+v", err)
	}

	r := &Response{}
	err = json.Unmarshal(b, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (c *clientImpl) Close() {
	_ = c.conn.Close()
}
