package main

import (
	"github.com/jacobsa/go-serial/serial"
	"github.com/spf13/cobra"
)

var serialOptions = serial.OpenOptions{
	MinimumReadSize: 8,
	ParityMode:      serial.PARITY_EVEN,
}

var jsonOuput bool

var rootCmd = &cobra.Command{
	Use: "smartmeter",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&serialOptions.PortName, "serial-port", "/dev/ttyUSB0", "device name to read packets from")
	rootCmd.PersistentFlags().UintVar(&serialOptions.BaudRate, "baud-rate", 9600, "baud rate")
	rootCmd.PersistentFlags().UintVar(&serialOptions.DataBits, "data-bits", 7, "data bits")
	rootCmd.PersistentFlags().UintVar(&serialOptions.StopBits, "stop-bits", 1, "stop bits")
	rootCmd.PersistentFlags().BoolVar(&jsonOuput, "json", false, "output as JSON")
}
