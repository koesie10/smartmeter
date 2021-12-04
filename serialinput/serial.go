package serialinput

import (
	"fmt"
	"github.com/jacobsa/go-serial/serial"
	"io"
	"strings"
)

type SerialOptions struct {
	Port       string     `env:"SERIAL_PORT" flag:"port" desc:"device name to read packets from"`
	BaudRate   uint       `env:"SERIAL_BAUD_RATE" flag:"baud-rate" desc:"baud rate"`
	DataBits   uint       `env:"SERIAL_DATA_BITS" flag:"data-bits" desc:"data bits"`
	StopBits   uint       `env:"SERIAL_STOP_BITS" flag:"stop-bits" desc:"stop bits"`
	ParityMode ParityMode `env:"SERIAL_PARITY_MODE" flag:"parity-mode" desc:"parity mode"`
}

func OpenSerial(opts *SerialOptions) (io.ReadCloser, error) {
	openOptions := serial.OpenOptions{
		PortName:        opts.Port,
		BaudRate:        opts.BaudRate,
		DataBits:        opts.DataBits,
		StopBits:        opts.StopBits,
		ParityMode:      serial.ParityMode(opts.ParityMode),
		MinimumReadSize: 1,
	}

	port, err := serial.Open(openOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port %s: %w", openOptions.PortName, err)
	}

	return port, nil
}

type ParityMode serial.ParityMode

func (m *ParityMode) String() string {
	switch serial.ParityMode(*m) {
	case serial.PARITY_NONE:
		return "N"
	case serial.PARITY_ODD:
		return "O"
	case serial.PARITY_EVEN:
		return "E"
	}
	panic("invalid parity")
}

func (m *ParityMode) Set(str string) error {
	if len(str) < 1 {
		return fmt.Errorf("invalid parity: empty")
	}

	switch strings.ToUpper(str)[0] {
	case 'N':
		*m = ParityMode(serial.PARITY_NONE)
	case 'O':
		*m = ParityMode(serial.PARITY_ODD)
	case 'E':
		*m = ParityMode(serial.PARITY_EVEN)
	default:
		return fmt.Errorf("unknown parity mode %q, expected N, O or E", str)
	}

	return nil
}

func (m *ParityMode) Type() string {
	return "string"
}
