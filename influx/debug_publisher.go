package influx

import (
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/koesie10/smartmeter/smartmeter"
	"time"
)

var _ smartmeter.Publisher = (*publisher)(nil)

type debugPublisher struct {
	options DebugPublisherOptions

	tags map[string]string
}

func NewDebugPublisher(options DebugPublisherOptions) (smartmeter.Publisher, error) {
	return &debugPublisher{
		options: options,
		tags:    map[string]string{},
	}, nil
}

type DebugPublisherOptions struct {
	ElectricityMeasurementName string `env:"INFLUX_ELECTRICITY_MEASUREMENT_NAME" flag:"electricity-measurement-name" desc:"InfluxDB electricity measurement name"`
	PhaseMeasurementName       string `env:"INFLUX_PHASE_MEASUREMENT_NAME" flag:"phase-measurement-name" desc:"InfluxDB phase measurement name"`
	GasMeasurementName         string `env:"INFLUX_GAS_PHASE_NAME" flag:"gas-measurement-name" desc:"InfluxDB gas measurement name"`
}

func (p *debugPublisher) Publish(packet *smartmeter.P1Packet) error {
	electricityPoint, err := NewElectricityPoint(time.Now(), packet, p.options.ElectricityMeasurementName, p.tags)
	if err != nil {
		return fmt.Errorf("failed to create electricity point: %w", err)
	}

	fmt.Printf("INFLUX DEBUG: %s", write.PointToLineProtocol(electricityPoint, time.Millisecond))

	for i := range packet.Electricity.Phases {
		phasePoint, err := NewPhasePoint(time.Now(), packet, i, p.options.PhaseMeasurementName, p.tags)
		if err != nil {
			return fmt.Errorf("failed to create phase point: %w", err)
		}

		fmt.Printf("INFLUX DEBUG: %s", write.PointToLineProtocol(phasePoint, time.Millisecond))
	}

	gasPoint, err := NewGasPoint(packet, p.options.GasMeasurementName, p.tags)
	if err != nil {
		return fmt.Errorf("failed to create gas point: %w", err)
	}

	fmt.Printf("INFLUX DEBUG: %s", write.PointToLineProtocol(gasPoint, time.Millisecond))

	return nil
}

func (p *debugPublisher) Close() error {
	return nil
}
