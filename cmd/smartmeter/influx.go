package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/koesie10/smartmeter/influx"
	"github.com/koesie10/smartmeter/smartmeter"
	"github.com/spf13/cobra"
)

var influxOptions = client.HTTPConfig{}

var additionalInfluxOptions = struct {
	Database        string
	RetentionPolicy string

	ElectricityMeasurement string
	PhaseMeasurement       string
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

		c, err := client.NewHTTPClient(influxOptions)
		if err != nil {
			return fmt.Errorf("failed to connect to InfluxDB: %v", err)
		}
		defer c.Close()

		if !additionalInfluxOptions.DisableUpload {
			if _, _, err := c.Ping(5 * time.Second); err != nil {
				return fmt.Errorf("failed to ping InfluxDB: %v", err)
			}
		}

		port, err := OpenPort()
		if err != nil {
			return fmt.Errorf("failed to open port: %v", err)
		}
		defer port.Close()

		sm, err := smartmeter.New(port)
		if err != nil {
			return fmt.Errorf("failed to open smart meter: %v", err)
		}

		for {
			packet, err := sm.Read()
			if err != nil {
				if _, ok := err.(*smartmeter.ParseError); !ok {
					return fmt.Errorf("failed to read packet: %v", err)
				}
				log.Println(err)
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

			pps := make([]*client.Point, 0, len(packet.Electricity.Phases))
			for i := range packet.Electricity.Phases {
				pp, err := influx.NewPhasePoint(time.Now(), packet, i, additionalInfluxOptions.PhaseMeasurement, tags)
				if err != nil {
					log.Println(err)
				}

				pps = append(pps, pp)
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
			bp.AddPoints(pps)
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
	influxCmd.Flags().StringVar(&additionalInfluxOptions.PhaseMeasurement, "influx-phase-measurement", "smartmeter_phase", "InfluxDB measurement for phase")
	influxCmd.Flags().StringVar(&additionalInfluxOptions.GasMeasurement, "influx-gas-measurement", "smartmeter_gas", "InfluxDB measurement for gas")

	influxCmd.Flags().StringSliceVar(&additionalInfluxOptions.Tags, "influx-tags", []string{}, "InfluxDB tags in key=value format")

	influxCmd.Flags().BoolVar(&additionalInfluxOptions.DisableUpload, "disable-upload", false, "if upload is disabled, then all points will be written to stdout")
}
