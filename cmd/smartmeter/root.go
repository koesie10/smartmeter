package main

import (
	"fmt"
	"strings"

	"github.com/jacobsa/go-serial/serial"
	"github.com/spf13/cobra"
)

var parityMode = ParityMode(serial.PARITY_EVEN)

var serialOptions = serial.OpenOptions{
	MinimumReadSize: 1,
}

var jsonOuput bool

var rootCmd = &cobra.Command{
	Use: "smartmeter",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		serialOptions.ParityMode = serial.ParityMode(parityMode)

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&serialOptions.PortName, "serial-port", "/dev/ttyUSB0", "device name to read packets from")
	rootCmd.PersistentFlags().UintVar(&serialOptions.BaudRate, "baud-rate", 9600, "baud rate")
	rootCmd.PersistentFlags().UintVar(&serialOptions.DataBits, "data-bits", 7, "data bits")
	rootCmd.PersistentFlags().UintVar(&serialOptions.StopBits, "stop-bits", 1, "stop bits")
	rootCmd.PersistentFlags().Var(&parityMode, "parity-mode", "parity mode")
	rootCmd.PersistentFlags().BoolVar(&jsonOuput, "json", false, "output as JSON")
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
