package main

import (
	"github.com/jacobsa/go-serial/serial"
	"github.com/koesie10/pflagenv"
	"github.com/koesie10/smartmeter/serialinput"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"log"
	"time"
)

var config = struct {
	serialinput.Options `env:",squish"`
}{
	Options: serialinput.Options{
		InputType: serialinput.SerialPort,

		Serial: &serialinput.SerialOptions{
			Port:       "/dev/ttyUSB0",
			BaudRate:   115200,
			DataBits:   8,
			StopBits:   1,
			ParityMode: serialinput.ParityMode(serial.PARITY_NONE),
		},

		File: &serialinput.FileOptions{
			Filename: "smartmeter/test/esmr50.txt",

			RepeatDelay: 1 * time.Second,
		},

		Network: &serialinput.NetworkOptions{
			Type:    "tcp",
			Address: "127.0.0.1:8888",
		},
	},
}

var logger, _ = zap.NewDevelopment()

var rootCmd = &cobra.Command{
	Use: "smartmeter",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := pflagenv.Parse(&config); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	if err := pflagenv.Setup(rootCmd.PersistentFlags(), &config); err != nil {
		log.Fatal(err)
	}
}
