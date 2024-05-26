package serialinput

import (
	"fmt"
	"io"
	"net"
	"time"
)

type NetworkOptions struct {
	Type        string        `env:"NETWORK_TYPE" flag:"type" desc:"network type"`
	Address     string        `env:"NETWORK_ADDRESS" flag:"address" desc:"network address"`
	DialTimeout time.Duration `env:"NETWORK_DIAL_TIMEOUT" flag:"dial-timeout" desc:"network dial timeout"`
	ReadTimeout time.Duration `env:"NETWORK_READ_TIMEOUT" flag:"read-timeout" desc:"network read timeout"`
}

func OpenNetwork(opts *NetworkOptions) (io.ReadCloser, error) {
	conn, err := net.DialTimeout(opts.Type, opts.Address, opts.DialTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to dial to network %s %s: %w", opts.Type, opts.Address, err)
	}

	return &timeoutReader{
		Conn:        conn,
		readTimeout: opts.ReadTimeout,
	}, nil
}

type timeoutReader struct {
	net.Conn
	readTimeout time.Duration
}

func (r *timeoutReader) Read(p []byte) (n int, err error) {
	if r.readTimeout > 0 {
		if err := r.Conn.SetReadDeadline(time.Now().Add(r.readTimeout)); err != nil {
			return 0, fmt.Errorf("failed to set read deadline: %w", err)
		}
	}
	return r.Conn.Read(p)
}
