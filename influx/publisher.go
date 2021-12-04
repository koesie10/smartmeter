package influx

import (
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/koesie10/smartmeter/smartmeter"
	"strings"
	"time"
)

var _ smartmeter.Publisher = (*publisher)(nil)

type publisher struct {
	client   influxdb2.Client
	writeAPI api.WriteAPI

	options PublisherOptions
	tags    map[string]string
}

func NewPublisher(options PublisherOptions) (smartmeter.Publisher, error) {
	influxOptions := influxdb2.DefaultOptions()
	influxOptions.SetPrecision(time.Second)
	client := influxdb2.NewClientWithOptions(options.Addr, options.AuthToken, influxOptions)

	writeAPI := client.WriteAPI(options.Organization, options.Bucket)

	tags := make(map[string]string)

	for _, v := range options.Tags {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid tag %q", v)
		}

		tags[parts[0]] = parts[1]
	}

	return &publisher{
		client:   client,
		writeAPI: writeAPI,
		options:  options,
		tags:     tags,
	}, nil
}

type PublisherOptions struct {
	Addr         string `env:"INFLUX_ADDR" flag:"addr" desc:"InfluxDB HTTP address, set empty to disable"`
	AuthToken    string `env:"INFLUX_AUTH_TOKEN" flag:"auth-token" desc:"InfluxDB auth token, use username:password for InfluxDB 1.8"`
	Organization string `env:"INFLUX_ORGANIZATION" flag:"organization" desc:"InfluxDB organization, do not set if using InfluxDB 1.8"`
	Bucket       string `env:"INFLUX_BUCKET" flag:"bucket" desc:"InfluxDB bucket, set to database/retention-policy or database for InfluxDB 1.8"`

	ElectricityMeasurementName string `env:"INFLUX_ELECTRICITY_MEASUREMENT_NAME" flag:"electricity-measurement-name" desc:"InfluxDB electricity measurement name"`
	PhaseMeasurementName       string `env:"INFLUX_PHASE_MEASUREMENT_NAME" flag:"phase-measurement-name" desc:"InfluxDB phase measurement name"`
	GasMeasurementName         string `env:"INFLUX_GAS_PHASE_NAME" flag:"gas-measurement-name" desc:"InfluxDB gas measurement name"`

	Tags []string `env:"INFLUX_TAGS" flag:"tags" desc:"InfluxDB tags in key=value format"`

	Timeout time.Duration `env:"INFLUX_TIMEOUT" flag:"timeout" desc:"InfluxDB timeout"`
}

func (p *publisher) Publish(packet *smartmeter.P1Packet) error {
	electricityPoint, err := NewElectricityPoint(time.Now(), packet, p.options.ElectricityMeasurementName, p.tags)
	if err != nil {
		return fmt.Errorf("failed to create electricity point: %w", err)
	}

	p.writeAPI.WritePoint(electricityPoint)

	for i := range packet.Electricity.Phases {
		phasePoint, err := NewPhasePoint(time.Now(), packet, i, p.options.PhaseMeasurementName, p.tags)
		if err != nil {
			return fmt.Errorf("failed to create phase point: %w", err)
		}

		p.writeAPI.WritePoint(phasePoint)
	}

	gasPoint, err := NewGasPoint(packet, p.options.GasMeasurementName, p.tags)
	if err != nil {
		return fmt.Errorf("failed to create gas point: %w", err)
	}

	p.writeAPI.WritePoint(gasPoint)

	return nil
}

func (p *publisher) Close() error {
	p.client.Close()

	return nil
}
