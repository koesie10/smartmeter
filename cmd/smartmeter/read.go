package main

import (
	"fmt"
	"github.com/koesie10/smartmeter/serialinput"
	"os"
	"text/tabwriter"
	"time"

	"github.com/koesie10/smartmeter/smartmeter"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "read a single P1 packet to stdout",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := serialinput.Open(&config.Options)
		if err != nil {
			return fmt.Errorf("failed to open port: %v", err)
		}
		defer port.Close()

		sm, err := smartmeter.New(port)
		if err != nil {
			return fmt.Errorf("failed to open smart meter: %v", err)
		}

		packet, err := sm.Read()
		if err != nil {
			return fmt.Errorf("failed to read packet: %v", err)
		}

		tw := tabwriter.NewWriter(os.Stdout, 10, 0, 2, ' ', tabwriter.AlignRight)
		fmt.Fprintln(tw, "Time\tTotal kWh Tariff 1 Consumed\tTotal kWh Tariff 2 consumed\tTotal gas consumed m^3\tCurrent consumption kW\tGas Measured At")
		fmt.Fprintf(tw, "%s\t%.3f\t%.3f\t%.3f\t%.3f\t%s", time.Now(), packet.Electricity.Tariffs[0].Consumed, packet.Electricity.Tariffs[1].Consumed, packet.Gas.Consumed, packet.Electricity.CurrentConsumed-packet.Electricity.CurrentProduced, packet.Gas.MeasuredAt)
		return tw.Flush()
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
}
