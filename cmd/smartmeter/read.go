package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/jacobsa/go-serial/serial"
	"github.com/koesie10/smartmeter/smartmeter"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "read a single P1 packet to stdout",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := serial.Open(serialOptions)
		if err != nil {
			return fmt.Errorf("failed to open serial port %s: %v", serialOptions.PortName, err)
		}

		sm, err := smartmeter.New(port)
		if err != nil {
			return fmt.Errorf("failed to open smart meter: %v", err)
		}

		packet, err := sm.Read()
		if err != nil {
			return fmt.Errorf("failed to read packet: %v", err)
		}

		if jsonOuput {
			data, err := json.Marshal(packet)
			if err != nil {
				return fmt.Errorf("failed to output JSON: %v", err)
			}

			fmt.Println(string(data))
			return nil
		}

		tw := tabwriter.NewWriter(os.Stdout, 10, 0, 2, ' ', tabwriter.AlignRight)
		fmt.Println(tw, "Time\tTotal kWh Tariff 1 Consumed\tTotal kWh Tariff 2 consumed\tTotal gas consumed\tCurrent kWh tariff\tGas Measured At")
		fmt.Fprintf(tw, "%s\t%.3f\t%.3f\t%.3f\t%d\t%s", time.Now(), packet.Electricity.Tariff1.Consumed, packet.Electricity.Tariff2.Consumed, packet.Gas.Consumed, packet.Gas.MeasuredAt)
		return tw.Flush()
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
}
