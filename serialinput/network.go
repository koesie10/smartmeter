package serialinput

import (
	"fmt"
	"io"
	"net"
	"time"
)

type NetworkOptions struct {
	Type    string `env:"NETWORK_TYPE" flag:"type" desc:"network type"`
	Address string `env:"NETWORK_ADDRESS" flag:"address" desc:"network address"`
}

func OpenNetwork(opts *NetworkOptions) (io.ReadCloser, error) {
	conn, err := net.DialTimeout(opts.Type, opts.Address, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to dial to network %s %s: %w", opts.Type, opts.Address, err)
	}

	return conn, nil
}
