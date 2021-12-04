package serialinput

import (
	"fmt"
	"io"
	"strings"
)

type Options struct {
	InputType InputType `env:"INPUT_TYPE" flag:"input-type" desc:"input type to read from"`

	Serial *SerialOptions `env:",squish"`

	File *FileOptions `env:",squish"`

	Network *NetworkOptions `env:",squish"`
}

func Open(opts *Options) (io.ReadCloser, error) {
	switch opts.InputType {
	case SerialPort:
		return OpenSerial(opts.Serial)
	case File:
		return OpenFile(opts.File)
	case Network:
		return OpenNetwork(opts.Network)
	}

	return nil, fmt.Errorf("unknown input type %v", opts.InputType)
}

type InputType int

const (
	SerialPort InputType = iota
	File
	Network
)

func (m *InputType) String() string {
	switch InputType(*m) {
	case SerialPort:
		return "serial"
	case File:
		return "file"
	case Network:
		return "network"
	}
	panic("invalid input type")
}

func (m *InputType) Set(str string) error {
	if len(str) < 1 {
		return fmt.Errorf("invalid input type: empty")
	}

	switch strings.ToLower(str) {
	case "serial":
		*m = SerialPort
	case "file":
		*m = File
	case "network":
		*m = Network
	default:
		return fmt.Errorf("unknown input type %q", str)
	}

	return nil
}

func (m *InputType) Type() string {
	return "string"
}
