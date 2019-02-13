package main

import (
	"fmt"
	"github.com/koesie10/smartmeter/smartmeter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var promOptions = struct {
	Addr string
}{}

var prometheusCmd = &cobra.Command{
	Use:   "prometheus",
	Short: "Start a web server and serve the data of all incoming P1 packets on /metrics",
	RunE: func(cmd *cobra.Command, args []string) error {
		tags := make(map[string]string)

		for _, v := range additionalInfluxOptions.Tags {
			parts := strings.SplitN(v, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid tag %q", v)
			}

			tags[parts[0]] = parts[1]
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

		registry := prometheus.NewRegistry()

		threshold := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:      "threshold",
			Help:      "Actual electricity threshold in the unit specified by the tags",
			Subsystem: "electricity",
			Namespace: "smartmeter",
		}, []string{"unit"})

		currentConsumed := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:      "current_consumed",
			Help:      "Actual electricity power delivered in kW",
			Subsystem: "electricity",
			Namespace: "smartmeter",
		})
		currentProduced := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:      "current_produced",
			Help:      "Actual electricity power produced in kW",
			Subsystem: "electricity",
			Namespace: "smartmeter",
		})

		numberOfPowerFailures := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:      "number_of_power_failures",
			Help:      "Number of power failures in any phase",
			Subsystem: "electricity",
			Namespace: "smartmeter",
		})
		numberOfLongPowerFailures := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:      "number_of_long_power_failures",
			Help:      "Number of long power failures in any phase",
			Subsystem: "electricity",
			Namespace: "smartmeter",
		})

		registry.MustRegister(threshold, currentConsumed, currentProduced, numberOfPowerFailures, numberOfLongPowerFailures)

		tariffConsumed := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:      "tariff_consumed",
			Help:      "Electricity delivered to client in kWh",
			Subsystem: "electricity",
			Namespace: "smartmeter",
		}, []string{"tariff"})
		tariffProduced := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:      "tariff_produced",
			Help:      "Electricity delivered by client in kWh",
			Subsystem: "electricity",
			Namespace: "smartmeter",
		}, []string{"tariff"})

		registry.MustRegister(tariffConsumed, tariffProduced)

		numberOfVoltageSags := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:      "number_of_voltage_sags",
			Help:      "Number of voltage sags in this phase",
			Subsystem: "electricity",
			Namespace: "smartmeter",
		}, []string{"phase"})
		numberOfVoltageSwells := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:      "number_of_voltage_swells",
			Help:      "Number of voltage swells in this phase",
			Subsystem: "electricity",
			Namespace: "smartmeter",
		}, []string{"phase"})

		instantaneousVoltage := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:      "instantaneous_voltage",
			Help:      "Instantaneous voltage in this phase in V",
			Subsystem: "electricity",
			Namespace: "smartmeter",
		}, []string{"phase"})
		instantaneousCurrent := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:      "instantaneous_current",
			Help:      "Instantaneous current in this phase in A",
			Subsystem: "electricity",
			Namespace: "smartmeter",
		}, []string{"phase"})
		instantaneousActivePositivePower := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:      "instantaneous_active_positive_power",
			Help:      "Instantaneous active power (+P) in this phase in kW",
			Subsystem: "electricity",
			Namespace: "smartmeter",
		}, []string{"phase"})
		instantaneousActiveNegativePower := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:      "instantaneous_active_negative_power",
			Help:      "Instantaneous active power (-P) in this phase in kW",
			Subsystem: "electricity",
			Namespace: "smartmeter",
		}, []string{"phase"})

		registry.MustRegister(numberOfVoltageSags, numberOfVoltageSwells, instantaneousVoltage)
		registry.MustRegister(instantaneousCurrent, instantaneousActivePositivePower, instantaneousActiveNegativePower)

		gasConsumed := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:      "consumed",
			Help:      "Gas delivered to client in m^3",
			Subsystem: "gas",
			Namespace: "smartmeter",
		})

		registry.MustRegister(gasConsumed)

		go func() {
			http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

			log.Fatal(http.ListenAndServe(promOptions.Addr, nil))
		}()

		for {
			p, err := sm.Read()
			if err != nil {
				return fmt.Errorf("failed to read packet: %v", err)
			}

			threshold.WithLabelValues(p.Electricity.ThresholdUnit).Set(p.Electricity.Threshold)

			currentConsumed.Set(p.Electricity.CurrentConsumed)
			currentProduced.Set(p.Electricity.CurrentProduced)

			numberOfPowerFailures.Set(float64(p.Electricity.NumberOfPowerFailures))
			numberOfLongPowerFailures.Set(float64(p.Electricity.NumberOfLongPowerFailures))

			for i, v := range p.Electricity.Tariffs {
				// Our tariffs are 0-indexed in the slice, while they are named in a 1-index fashion.
				tariff := strconv.Itoa(i + 1)

				tariffConsumed.WithLabelValues(tariff).Set(v.Consumed)
				tariffProduced.WithLabelValues(tariff).Set(v.Produced)
			}

			for i, v := range p.Electricity.Phases {
				// Our phases are 0-indexed in the slice, while they are named in a 1-index fashion.
				phase := strconv.Itoa(i + 1)

				numberOfVoltageSags.WithLabelValues(phase).Set(float64(v.NumberOfVoltageSags))
				numberOfVoltageSwells.WithLabelValues(phase).Set(float64(v.NumberOfVoltageSwells))

				instantaneousVoltage.WithLabelValues(phase).Set(v.InstantaneousVoltage)
				instantaneousCurrent.WithLabelValues(phase).Set(v.InstantaneousCurrent)
				instantaneousActivePositivePower.WithLabelValues(phase).Set(v.InstantaneousActivePositivePower)
				instantaneousActiveNegativePower.WithLabelValues(phase).Set(v.InstantaneousActiveNegativePower)
			}

			gasConsumed.Set(p.Gas.Consumed)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(prometheusCmd)
	prometheusCmd.Flags().StringVar(&promOptions.Addr, "addr", ":8888", "Web server address")
}

func copyTags(tags map[string]string) map[string]string {
	if len(tags) == 0 {
		return tags
	}
	result := make(map[string]string, len(tags))
	for k, v := range tags {
		result[k] = v
	}
	return result
}
