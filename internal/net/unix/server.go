// Package unix communicate internal access
package unix

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/hatappi/go-kit/log"
	"github.com/hatappi/gomodoro/internal/timer"
)

// Response represents unix server response
type Response struct {
	RemainSec int
}

// GetRemain get remain string
func (r *Response) GetRemain() string {
	if r.RemainSec == 0 {
		return "00:00"
	}
	min := r.RemainSec / 60
	sec := r.RemainSec % 60
	return fmt.Sprintf("%02d:%02d", min, sec)
}

// Server represents server
type Server interface {
	Serve(context.Context)
	Close()
}

type serverImpl struct {
	listener net.Listener
	timer    timer.Timer
}

// NewServer initialize Server
func NewServer(socketPath string, timer timer.Timer) (Server, error) {
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}

	return &serverImpl{
		listener: listener,
		timer:    timer,
	}, nil
}

// Serve start unix domain socket server
func (c *serverImpl) Serve(ctx context.Context) {
	for {
		conn, err := c.listener.Accept()
		if err != nil {
			log.FromContext(ctx).Error(err, "failed to accpect")
			return
		}
		log.FromContext(ctx).V(1).Info("accept request")

		go func() {
			defer func() {
				_ = conn.Close()
			}()

			rs := c.timer.GetRemainSec()

			r := &Response{
				RemainSec: rs,
			}

			b, err := json.Marshal(r)
			if err != nil {
				log.FromContext(ctx).Error(err, "faield to marshal Response")
				return
			}

			if _, err = conn.Write(b); err != nil {
				log.FromContext(ctx).Error(err, "failed to write response")
			}
		}()
	}
}

func (c *serverImpl) Close() {
	_ = c.listener.Close()
}
