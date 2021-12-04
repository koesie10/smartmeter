package main

import (
	"encoding/json"
	"fmt"
	"github.com/koesie10/pflagenv"
	"github.com/koesie10/smartmeter/influx"
	"github.com/koesie10/smartmeter/prometheus"
	"github.com/koesie10/smartmeter/smartmeter"
	"github.com/spf13/cobra"
	"log"
)

var publishConfig = struct {
	Influx     influx.PublisherOptions     `env:",squash"`
	Prometheus prometheus.PublisherOptions `env:",squash"`

	EnableJSONDebug   bool `env:"ENABLE_JSON_DEBUG" flag:"enable-json-debug" desc:"enable json debug output"`
	EnableInfluxDebug bool `env:"ENABLE_INFLUX_DEBUG" flag:"enable-influx-debug" desc:"enable influx debug output"`
}{
	Influx: influx.PublisherOptions{
		Addr:   "http://localhost:8086",
		Bucket: "smartmeter",

		ElectricityMeasurementName: "smartmeter_electricity",
		PhaseMeasurementName:       "smartmeter_phase",
		GasMeasurementName:         "smartmeter_gas",
	},

	Prometheus: prometheus.PublisherOptions{
		Addr: ":8888",
	},
}

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish P1 packets to various publishers, including InfluxDB and Prometheus",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := pflagenv.Parse(&publishConfig); err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPublish()
	},
}

func runPublish() error {
	var publishers []smartmeter.Publisher

	if publishConfig.Influx.Addr != "" {
		publisher, err := influx.NewPublisher(publishConfig.Influx)
		if err != nil {
			return fmt.Errorf("failed to create InfluxDB publisher: %w", err)
		}
		defer publisher.Close()
		publishers = append(publishers, publisher)
	}

	if publishConfig.Prometheus.Addr != "" {
		publisher, err := prometheus.NewPublisher(publishConfig.Prometheus)
		if err != nil {
			return fmt.Errorf("failed to create Prometheus publisher: %w", err)
		}
		defer publisher.Close()
		publishers = append(publishers, publisher)
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
			continue
		}

		if jsonOuput {
			data, err := json.Marshal(packet)
			if err != nil {
				return fmt.Errorf("failed to output JSON: %v", err)
			}

			fmt.Println(string(data))
		}

		for _, publisher := range publishers {
			if err := publisher.Publish(packet); err != nil {
				log.Println(err)
			}
		}
	}
}

func init() {
	rootCmd.AddCommand(publishCmd)

	if err := pflagenv.Setup(publishCmd.Flags(), &publishConfig); err != nil {
		log.Fatal(err)
	}
}
