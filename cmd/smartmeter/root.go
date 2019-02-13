package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/jacobsa/go-serial/serial"
	"github.com/spf13/cobra"
)

var inputType = SerialPort
var parityMode = ParityMode(serial.PARITY_EVEN)

var serialOptions = serial.OpenOptions{
	MinimumReadSize: 1,
}

var fileOptions = struct {
	Filename string
}{}

var networkOptions = struct {
	Network string
	Address string
}{}

var jsonOuput bool

var rootCmd = &cobra.Command{
	Use: "smartmeter",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		serialOptions.ParityMode = serial.ParityMode(parityMode)

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().Var(&inputType, "input-type", "input type to read from")

	rootCmd.PersistentFlags().StringVar(&serialOptions.PortName, "serial-port", "/dev/ttyUSB0", "device name to read packets from")
	rootCmd.PersistentFlags().UintVar(&serialOptions.BaudRate, "baud-rate", 9600, "baud rate")
	rootCmd.PersistentFlags().UintVar(&serialOptions.DataBits, "data-bits", 7, "data bits")
	rootCmd.PersistentFlags().UintVar(&serialOptions.StopBits, "stop-bits", 1, "stop bits")
	rootCmd.PersistentFlags().Var(&parityMode, "parity-mode", "parity mode")

	rootCmd.PersistentFlags().StringVar(&fileOptions.Filename, "filename", "smartmeter/test/esmr50.txt", "filename to read from")

	rootCmd.PersistentFlags().StringVar(&networkOptions.Network, "network-type", "tcp", "network type")
	rootCmd.PersistentFlags().StringVar(&networkOptions.Address, "network-address", "127.0.0.1:8888", "network address")

	rootCmd.PersistentFlags().BoolVar(&jsonOuput, "json", false, "output as JSON")
}

func OpenPort() (io.ReadCloser, error) {
	switch inputType {
	case SerialPort:
		port, err := serial.Open(serialOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to open serial port %s: %v", serialOptions.PortName, err)
		}

		return port, nil
	case File:
		file, err := os.Open(fileOptions.Filename)
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %v", fileOptions.Filename, err)
		}

		return file, nil
	case Network:
		conn, err := net.DialTimeout(networkOptions.Network, networkOptions.Address, 10 * time.Second)
		if err != nil {
			return nil, fmt.Errorf("failed to dial to network %s %s: %v", networkOptions.Network, networkOptions.Address, err)
		}

		return conn, nil
	}

	return nil, fmt.Errorf("unknown input type %v", inputType)
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
