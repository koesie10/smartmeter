package prometheus

import (
	"fmt"
	"github.com/koesie10/smartmeter/smartmeter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net"
	"net/http"
	"strconv"
)

var _ smartmeter.Publisher = (*publisher)(nil)

type publisher struct {
	options PublisherOptions

	server *http.Server

	threshold       *prometheus.GaugeVec
	currentConsumed prometheus.Gauge
	currentProduced prometheus.Gauge

	numberOfPowerFailures     prometheus.Gauge
	numberOfLongPowerFailures prometheus.Gauge

	tariffConsumed *prometheus.GaugeVec
	tariffProduced *prometheus.GaugeVec

	numberOfVoltageSags   *prometheus.GaugeVec
	numberOfVoltageSwells *prometheus.GaugeVec

	instantaneousVoltage             *prometheus.GaugeVec
	instantaneousCurrent             *prometheus.GaugeVec
	instantaneousActivePositivePower *prometheus.GaugeVec
	instantaneousActiveNegativePower *prometheus.GaugeVec

	gasConsumed prometheus.Gauge
}

func NewPublisher(options PublisherOptions) (smartmeter.Publisher, error) {
	p := &publisher{
		options: options,
	}

	registry := prometheus.NewRegistry()

	p.threshold = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "threshold",
		Help:      "Actual electricity threshold in the unit specified by the tags",
		Subsystem: "electricity",
		Namespace: "smartmeter",
	}, []string{"unit"})

	p.currentConsumed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "current_consumed",
		Help:      "Actual electricity power delivered in kW",
		Subsystem: "electricity",
		Namespace: "smartmeter",
	})
	p.currentProduced = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "current_produced",
		Help:      "Actual electricity power produced in kW",
		Subsystem: "electricity",
		Namespace: "smartmeter",
	})

	p.numberOfPowerFailures = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "number_of_power_failures",
		Help:      "Number of power failures in any phase",
		Subsystem: "electricity",
		Namespace: "smartmeter",
	})
	p.numberOfLongPowerFailures = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "number_of_long_power_failures",
		Help:      "Number of long power failures in any phase",
		Subsystem: "electricity",
		Namespace: "smartmeter",
	})

	registry.MustRegister(p.threshold, p.currentConsumed, p.currentProduced, p.numberOfPowerFailures, p.numberOfLongPowerFailures)

	p.tariffConsumed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "tariff_consumed",
		Help:      "Electricity delivered to client in kWh",
		Subsystem: "electricity",
		Namespace: "smartmeter",
	}, []string{"tariff"})
	p.tariffProduced = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "tariff_produced",
		Help:      "Electricity delivered by client in kWh",
		Subsystem: "electricity",
		Namespace: "smartmeter",
	}, []string{"tariff"})

	registry.MustRegister(p.tariffConsumed, p.tariffProduced)

	p.numberOfVoltageSags = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "number_of_voltage_sags",
		Help:      "Number of voltage sags in this phase",
		Subsystem: "electricity",
		Namespace: "smartmeter",
	}, []string{"phase"})
	p.numberOfVoltageSwells = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "number_of_voltage_swells",
		Help:      "Number of voltage swells in this phase",
		Subsystem: "electricity",
		Namespace: "smartmeter",
	}, []string{"phase"})

	p.instantaneousVoltage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "instantaneous_voltage",
		Help:      "Instantaneous voltage in this phase in V",
		Subsystem: "electricity",
		Namespace: "smartmeter",
	}, []string{"phase"})
	p.instantaneousCurrent = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "instantaneous_current",
		Help:      "Instantaneous current in this phase in A",
		Subsystem: "electricity",
		Namespace: "smartmeter",
	}, []string{"phase"})
	p.instantaneousActivePositivePower = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "instantaneous_active_positive_power",
		Help:      "Instantaneous active power (+P) in this phase in kW",
		Subsystem: "electricity",
		Namespace: "smartmeter",
	}, []string{"phase"})
	p.instantaneousActiveNegativePower = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "instantaneous_active_negative_power",
		Help:      "Instantaneous active power (-P) in this phase in kW",
		Subsystem: "electricity",
		Namespace: "smartmeter",
	}, []string{"phase"})

	registry.MustRegister(p.numberOfVoltageSags, p.numberOfVoltageSwells, p.instantaneousVoltage)
	registry.MustRegister(p.instantaneousCurrent, p.instantaneousActivePositivePower, p.instantaneousActiveNegativePower)

	p.gasConsumed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:      "consumed",
		Help:      "Gas delivered to client in m^3",
		Subsystem: "gas",
		Namespace: "smartmeter",
	})

	registry.MustRegister(p.gasConsumed)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	p.server = &http.Server{
		Addr:    options.Addr,
		Handler: mux,
	}

	lis, err := net.Listen("tcp", p.server.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to start listening: %w", err)
	}

	go func() {
		if err := p.server.Serve(lis); err != nil {
			log.Println(err)
		}
	}()

	return p, nil
}

type PublisherOptions struct {
	Addr string `env:"PROMETHEUS_ADDR" flag:"addr" desc:"Prometheus HTTP server address, set empty to disable"`
}

func (p *publisher) Publish(packet *smartmeter.P1Packet) error {
	p.threshold.WithLabelValues(packet.Electricity.ThresholdUnit).Set(packet.Electricity.Threshold)

	p.currentConsumed.Set(packet.Electricity.CurrentConsumed)
	p.currentProduced.Set(packet.Electricity.CurrentProduced)

	p.numberOfPowerFailures.Set(float64(packet.Electricity.NumberOfPowerFailures))
	p.numberOfLongPowerFailures.Set(float64(packet.Electricity.NumberOfLongPowerFailures))

	for i, v := range packet.Electricity.Tariffs {
		// Our tariffs are 0-indexed in the slice, while they are named in a 1-index fashion.
		tariff := strconv.Itoa(i + 1)

		p.tariffConsumed.WithLabelValues(tariff).Set(v.Consumed)
		p.tariffProduced.WithLabelValues(tariff).Set(v.Produced)
	}

	for i, v := range packet.Electricity.Phases {
		// Our phases are 0-indexed in the slice, while they are named in a 1-index fashion.
		phase := strconv.Itoa(i + 1)

		p.numberOfVoltageSags.WithLabelValues(phase).Set(float64(v.NumberOfVoltageSags))
		p.numberOfVoltageSwells.WithLabelValues(phase).Set(float64(v.NumberOfVoltageSwells))

		p.instantaneousVoltage.WithLabelValues(phase).Set(v.InstantaneousVoltage)
		p.instantaneousCurrent.WithLabelValues(phase).Set(v.InstantaneousCurrent)
		p.instantaneousActivePositivePower.WithLabelValues(phase).Set(v.InstantaneousActivePositivePower)
		p.instantaneousActiveNegativePower.WithLabelValues(phase).Set(v.InstantaneousActiveNegativePower)
	}

	p.gasConsumed.Set(packet.Gas.Consumed)

	return nil
}

func (p *publisher) Close() error {
	return p.server.Close()
}
