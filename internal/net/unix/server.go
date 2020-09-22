// Package unix communicate internal access
package unix

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/hatappi/go-kit/log"
	"github.com/hatappi/gomodoro/internal/timer"
	"go.uber.org/zap"
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
			log.FromContext(ctx).Warn("failed to accpect", zap.Error(err))
			return
		}
		log.FromContext(ctx).Debug("accept request")

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
				log.FromContext(ctx).Error("faield to marshal Response", zap.Error(err))
				return
			}

			if _, err = conn.Write(b); err != nil {
				log.FromContext(ctx).Error("failed to write response", zap.Error(err))
			}
		}()
	}
}

func (c *serverImpl) Close() {
	_ = c.listener.Close()
}
