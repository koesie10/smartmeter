package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/jacobsa/go-serial/serial"
	"github.com/koesie10/smartmeter/influx"
	"github.com/koesie10/smartmeter/smartmeter"
	"github.com/spf13/cobra"
)

var influxOptions = client.HTTPConfig{}

var additionalInfluxOptions = struct {
	Database        string
	RetentionPolicy string

	ElectricityMeasurement string
	GasMeasurement         string

	Tags []string

	DisableUpload bool
}{}

var influxCmd = &cobra.Command{
	Use:   "influx",
	Short: "send all incoming P1 packets to InfluxDB",
	RunE: func(cmd *cobra.Command, args []string) error {
		tags := make(map[string]string)

		for _, v := range additionalInfluxOptions.Tags {
			parts := strings.SplitN(v, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid tag %q", v)
			}

			tags[parts[0]] = parts[1]
		}

		var c client.Client

		if !additionalInfluxOptions.DisableUpload {
			c, err := client.NewHTTPClient(influxOptions)
			if err != nil {
				return fmt.Errorf("failed to connect to InfluxDB: %v", err)
			}
			defer c.Close()

			if _, _, err := c.Ping(5 * time.Second); err != nil {
				return fmt.Errorf("failed to ping InfluxDB: %v", err)
			}
		}

		port, err := serial.Open(serialOptions)
		if err != nil {
			return fmt.Errorf("failed to open serial port %s: %v", serialOptions.PortName, err)
		}
		defer port.Close()

		sm, err := smartmeter.New(port)
		if err != nil {
			return fmt.Errorf("failed to open smart meter: %v", err)
		}

		for {
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
			}

			ep, err := influx.NewElectricityPoint(time.Now(), packet, additionalInfluxOptions.ElectricityMeasurement, tags)
			if err != nil {
				log.Println(err)
				continue
			}

			gp, err := influx.NewGasPoint(packet, additionalInfluxOptions.GasMeasurement, tags)
			if err != nil {
				log.Println(err)
				continue
			}

			bp, err := client.NewBatchPoints(client.BatchPointsConfig{
				Database:        additionalInfluxOptions.Database,
				RetentionPolicy: additionalInfluxOptions.RetentionPolicy,
			})
			if err != nil {
				log.Println(err)
				continue
			}

			bp.AddPoint(ep)
			bp.AddPoint(gp)

			if additionalInfluxOptions.DisableUpload {
				for _, p := range bp.Points() {
					fmt.Println(p.PrecisionString("ns"))
				}
			} else {
				if err := c.Write(bp); err != nil {
					log.Println(err)
					continue
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(influxCmd)
	influxCmd.Flags().StringVar(&influxOptions.Addr, "influx-addr", "http://localhost:8086", "InfluxDB address")
	influxCmd.Flags().StringVar(&influxOptions.Username, "influx-username", "", "InfluxDB username")
	influxCmd.Flags().StringVar(&influxOptions.Password, "influx-password", "", "InfluxDB password")
	influxCmd.Flags().DurationVar(&influxOptions.Timeout, "influx-timeout", 10*time.Second, "InfluxDB timeout")

	influxCmd.Flags().StringVar(&additionalInfluxOptions.Database, "influx-database", "smartmeter", "InfluxDB database")
	influxCmd.Flags().StringVar(&additionalInfluxOptions.RetentionPolicy, "influx-retention-policy", "", "InfluxDB retention policy. Leave empty for default.")
	influxCmd.Flags().StringVar(&additionalInfluxOptions.ElectricityMeasurement, "influx-electricity-measurement", "smartmeter_electricity", "InfluxDB measurement for electricity")
	influxCmd.Flags().StringVar(&additionalInfluxOptions.GasMeasurement, "influx-gas-measurement", "smartmeter_gas", "InfluxDB measurement for gas")

	influxCmd.Flags().StringSliceVar(&additionalInfluxOptions.Tags, "influx-tags", []string{}, "InfluxDB tags in key=value format")

	influxCmd.Flags().BoolVar(&additionalInfluxOptions.DisableUpload, "disable-upload", false, "if upload is disabled, then all points will be written to stdout")
}
